package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/redis/go-redis/v9"
	slogctx "github.com/veqryn/slog-context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type Subscriber struct {
	lobbyService    LobbyServicer
	playerService   PlayerServicer
	roundService    RoundServicer
	logger          *slog.Logger
	handlerRegistry *HandlerRegistry
	websocket       Websocketer
	config          config.Config
	rules           templ.Component
	stateMachines   *statemachine.Manager
}

type Websocketer interface {
	Subscribe(ctx context.Context, id uuid.UUID) <-chan *redis.Message
	Publish(ctx context.Context, id uuid.UUID, msg []byte) error
	Close(id uuid.UUID) error
}

type message struct {
	MessageType string            `json:"message_type"`
	Headers     map[string]string `json:"HEADERS,omitempty"`
	Trace       *TraceContext     `json:"_trace,omitempty"`
	TestContext *TestContext      `json:"test_context,omitempty"`
}

type TraceContext struct {
	TraceID string `json:"traceId"`
	SpanID  string `json:"spanId"`
}

type TestContext struct {
	TestName string `json:"testName"`
	TestID   string `json:"testId"`
}

var errConnectionClosed = errors.New("connection closed")

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "write: connection refused") ||
		strings.Contains(errStr, "use of reserved op code")
}

func NewSubscriber(
	lobbyService LobbyServicer,
	playerService PlayerServicer,
	roundService RoundServicer,
	logger *slog.Logger,
	websocket Websocketer,
	config config.Config,
	rules templ.Component,
	shutdownCtx context.Context,
) *Subscriber {
	baseMiddleware := NewChain(
		RecoveryMiddleware(),
		LoggingMiddleware(),
		AuthMiddleware(),
	)

	registry := NewHandlerRegistry(baseMiddleware...)

	s := &Subscriber{
		lobbyService:    lobbyService,
		playerService:   playerService,
		roundService:    roundService,
		logger:          logger,
		handlerRegistry: registry,
		websocket:       websocket,
		config:          config,
		rules:           rules,
		stateMachines:   statemachine.NewManager(shutdownCtx, logger),
	}

	s.registerHandlers()
	return s
}

func (s *Subscriber) registerHandlers() {
	s.handlerRegistry.Register("create_room", WSHandlerAdapter(func() WSHandler { return &CreateRoom{} }))
	s.handlerRegistry.Register("join_lobby", WSHandlerAdapter(func() WSHandler { return &JoinLobby{} }))
	s.handlerRegistry.Register("start_game", WSHandlerAdapter(func() WSHandler { return &StartGame{} }))
	s.handlerRegistry.Register("kick_player", WSHandlerAdapter(func() WSHandler { return &KickPlayer{} }))
	s.handlerRegistry.Register(
		"update_player_nickname",
		WSHandlerAdapter(func() WSHandler { return &UpdateNickname{} }),
	)
	s.handlerRegistry.Register(
		"generate_new_avatar",
		WSHandlerAdapter(func() WSHandler { return &GenerateNewAvatar{} }),
	)
	s.handlerRegistry.Register(
		"toggle_player_is_ready",
		WSHandlerAdapter(func() WSHandler { return &TogglePlayerIsReady{} }),
	)

	s.handlerRegistry.Register("submit_answer", WSHandlerAdapter(func() WSHandler { return &SubmitAnswer{} }))
	s.handlerRegistry.Register(
		"toggle_answer_is_ready",
		WSHandlerAdapter(func() WSHandler { return &ToggleAnswerIsReady{} }),
	)
	s.handlerRegistry.Register("submit_vote", WSHandlerAdapter(func() WSHandler { return &SubmitVote{} }))
	s.handlerRegistry.Register(
		"toggle_voting_is_ready",
		WSHandlerAdapter(func() WSHandler { return &ToggleVotingIsReady{} }),
	)
}

func (s *Subscriber) Subscribe(r *http.Request, w http.ResponseWriter) (err error) {
	ctx := r.Context()
	ctx, cancel := context.WithCancel(ctx)
	start := time.Now()

	telemetry.IncrementActiveConnections()
	defer telemetry.DecrementActiveConnections()

	defer func() {
		latencyInSeconds := float64(time.Since(start).Seconds())
		err = telemetry.RecordConnectionDuration(ctx, latencyInSeconds)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to record connection time metric", slog.Any("error", err))
		}
	}()

	tracer := otel.Tracer("banterbus-websocket")
	ctx, span := tracer.Start(
		ctx,
		"websocket.subscribe",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.NetworkTransportKey.String("tcp"),
			semconv.NetworkTypeKey.String("ipv4"),
			semconv.NetworkProtocolName("websocket"),
			attribute.String("component", "websocket-subscriber"),
		),
	)

	locale := s.config.App.DefaultLocale.String()
	cookie, err := r.Cookie("locale")
	if err == nil {
		locale = cookie.Value
	}

	span.AddEvent("add_locale")
	ctx, err = ctxi18n.WithLocale(ctx, locale)
	if err != nil {
		span.AddEvent("failed_to_set_locale")
		s.logger.ErrorContext(
			ctx,
			"failed to set locale",
			slog.String("locale", locale),
			slog.Any("error", err),
		)

		ctx, err = ctxi18n.WithLocale(ctx, s.config.App.DefaultLocale.String())
		if err != nil {
			s.logger.ErrorContext(
				ctx,
				"failed to set locale to default",
				slog.String("locale", locale),
				slog.Any("error", err),
			)
		}
	}

	var component bytes.Buffer
	var playerID uuid.UUID

	cookie, err = r.Cookie("player_id")
	if err != nil {
		span.AddEvent("no_cookie_found")
		cookie = setPlayerIDCookie()
		http.SetCookie(w, cookie)
	} else {
		playerID, err = uuid.FromString(cookie.Value)
		if err != nil {
			cancel()
			return err
		}

		err = telemetry.IncrementReconnectionCount(ctx)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to increment reconnection count", slog.Any("error", err))
		}

		if s.config.App.AutoReconnect {
			component, err = s.Reconnect(ctx, playerID)
			if err != nil {
				s.logger.WarnContext(ctx, "failed to reconnect", slog.Any("error", err))
				cookie = setPlayerIDCookie()
				http.SetCookie(w, cookie)
			}
		}
	}

	playerID, err = uuid.FromString(cookie.Value)
	if err != nil {
		cancel()
		return err
	}

	span.SetAttributes(attribute.String("player_id", playerID.String()))
	err = s.playerService.UpdateLocale(ctx, playerID, locale)
	if err != nil {
		s.logger.WarnContext(
			ctx,
			"failed to update player locale",
			slog.Any("error", err),
			slog.String("locale", locale),
			slog.String("player_id", playerID.String()),
		)
	}

	h := ws.HTTPUpgrader{
		Header: w.Header(),
	}
	connection, _, _, err := h.Upgrade(r, w)
	if err != nil {
		span.AddEvent("connection_ws_upgrade_failed")
		err = telemetry.IncrementHandshakeFailures(ctx, err.Error())
		if err != nil {
			s.logger.WarnContext(ctx, "failed to increment handshake failure", slog.Any("error", err))
		}

		cancel()
		return err
	}
	span.AddEvent("connection_ws_upgraded")
	err = telemetry.IncrementSubscribers(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to increment counter", slog.Any("error", err))
	}

	subscribeCh := s.websocket.Subscribe(ctx, playerID)
	client := newClient(connection, playerID, subscribeCh)

	// INFO: Send the reconnection message to the client if they should reconnect.
	if component.Len() > 0 {
		err = s.websocket.Publish(ctx, playerID, component.Bytes())
	}

	defer func() {
		s.logger.InfoContext(ctx, "websocket connection closed",
			slog.String("player_id", playerID.String()),
		)
		cancel()

		err = s.lobbyService.HandlePlayerDisconnect(ctx, playerID)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to handle player disconnect",
				slog.String("player_id", playerID.String()),
				slog.Any("error", err))
		}

		err = s.websocket.Close(playerID)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to close websocket subscription", slog.Any("error", err))
		}
		err = connection.Close()
		if err != nil {
			s.logger.WarnContext(ctx, "failed to close connection", slog.Any("error", err))
		}
	}()

	span.SetStatus(codes.Ok, "subscribed_successfully")
	span.End()

	go s.handleMessages(ctx, cancel, client)

	for {
		select {

		case msg := <-client.messagesCh:
			start := time.Now()
			err = s.sendMessageWithRetry(ctx, connection, []byte(msg.Payload), playerID, 3)
			if err != nil {
				if isConnectionError(err) {
					s.logger.DebugContext(ctx, "client connection closed, stopping message loop",
						slog.String("player_id", playerID.String()))
					cancel()
					return nil
				}

				s.logger.ErrorContext(ctx, "failed to write message after retries", slog.Any("error", err))

				err = telemetry.IncrementMessageSentError(ctx)
				if err != nil {
					s.logger.WarnContext(ctx, "failed to increment message sent err", slog.Any("error", err))
				}
			} else {
				err = telemetry.IncrementMessageSent(ctx)
				if err != nil {
					s.logger.WarnContext(ctx, "failed to increment message sent", slog.Any("error", err))
				}

				err = telemetry.RecordMessageSendLatency(ctx, time.Since(start).Seconds())
				if err != nil {
					s.logger.WarnContext(ctx, "failed to record send latency", slog.Any("error", err))
				}
			}
		case <-ctx.Done():
			s.logger.DebugContext(ctx, "subscribe context done", slog.String("player_id", playerID.String()))
			cancel()
			return ctx.Err()
		}
	}
}

func setPlayerIDCookie() *http.Cookie {
	playerID, err := uuid.NewV7()
	if err != nil {
		// Fallback to NewV4 if NewV7 fails
		playerID = uuid.Must(uuid.NewV4())
	}

	cookie := &http.Cookie{
		Name:     "player_id",
		Value:    playerID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour),
	}
	return cookie
}

func (s *Subscriber) handleMessages(ctx context.Context, cancel context.CancelFunc, client *Client) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			tracer := otel.Tracer("banterbus-websocket")
			ctx, span := tracer.Start(
				ctx,
				"websocket.handle_message",
				trace.WithSpanKind(trace.SpanKindInternal),
				trace.WithAttributes(
					attribute.String("game.player_id", client.playerID.String()),
					attribute.String("component", "websocket-handler"),
				),
			)
			var err error
			ctx, err = s.handleMessage(ctx, client)
			span.End()
			if err != nil {
				if errors.Is(err, errConnectionClosed) || isConnectionError(err) {
					s.logger.DebugContext(ctx, "client connection closed, stopping message handler",
						slog.String("player_id", client.playerID.String()))
					cancel()
					return
				}

				s.logger.ErrorContext(
					ctx,
					"failed to handle message",
					slog.Any("error", err),
					slog.String("player_id", client.playerID.String()),
				)
				err := telemetry.IncrementMessageReceivedError(ctx)
				if err != nil {
					s.logger.WarnContext(
						ctx,
						"failed to increment message received error metric",
						slog.Any("error", err),
					)
				}

				telemetry.RecordWebSocketError(ctx)
			}
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, client *Client) (context.Context, error) {
	start := time.Now()
	messageStatus := "success"
	ctx = slogctx.Append(ctx, "player_id", client.playerID.String())

	hdr, r, err := wsutil.NextReader(client.connection, ws.StateServerSide)
	if err != nil {
		if err == io.EOF {
			return ctx, errConnectionClosed
		} else if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
			return ctx, errConnectionClosed
		} else if isConnectionError(err) {
			return ctx, errConnectionClosed
		}

		return ctx, fmt.Errorf("failed to get next message: %w", err)
	}

	if hdr.OpCode == ws.OpClose {
		return ctx, errConnectionClosed
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	data, err := io.ReadAll(r)
	if err != nil {
		return ctx, fmt.Errorf("failed to read message: %w", err)
	}

	return s.handleMessageData(ctx, client, data, start, messageStatus)
}

func (s *Subscriber) sendMessageWithRetry(
	ctx context.Context,
	connection net.Conn,
	data []byte,
	playerID uuid.UUID,
	maxRetries int,
) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := wsutil.WriteServerText(connection, data)
		if err == nil {
			if attempt > 0 {
				s.logger.DebugContext(ctx, "message sent successfully after retry",
					slog.String("player_id", playerID.String()),
					slog.Int("attempt", attempt+1))
			}
			return nil
		}

		if isConnectionError(err) {
			return err
		}

		lastErr = err
		if attempt < maxRetries {
			s.logger.WarnContext(ctx, "failed to send message, retrying",
				slog.String("player_id", playerID.String()),
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", maxRetries),
				slog.Any("error", err))

			time.Sleep(time.Millisecond * 50)
		}
	}

	s.logger.ErrorContext(ctx, "failed to send message after all retries",
		slog.String("player_id", playerID.String()),
		slog.Int("max_retries", maxRetries),
		slog.Any("error", lastErr))

	return lastErr
}

func (s *Subscriber) handleMessageData(
	ctx context.Context,
	client *Client,
	data []byte,
	start time.Time,
	messageStatus string,
) (context.Context, error) {

	// Debug log every inbound WebSocket request
	s.logger.DebugContext(ctx, "websocket request inbound",
		slog.String("player_id", client.playerID.String()),
		slog.Int("data_size", len(data)),
		slog.String("raw_data", string(data)),
	)

	var message message
	err := json.Unmarshal(data, &message)
	s.logger.DebugContext(ctx, "received message", slog.Any("message", message))

	if err != nil {
		messageStatus = "fail_unmarshal_message"
		return ctx, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	tracer := otel.Tracer("banterbus-websocket")
	var span trace.Span
	if message.Trace != nil && message.Trace.TraceID != "" {

		if len(message.Trace.TraceID) == 32 {
			traceID, err := trace.TraceIDFromHex(message.Trace.TraceID)
			if err == nil && len(message.Trace.SpanID) == 16 {
				spanID, err := trace.SpanIDFromHex(message.Trace.SpanID)
				if err == nil {
					spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID:    traceID,
						SpanID:     spanID,
						TraceFlags: trace.FlagsSampled,
					})
					bag := baggage.FromContext(ctx)
					ctxWithSpan := trace.ContextWithSpanContext(ctx, spanCtx)
					ctx = baggage.ContextWithBaggage(ctxWithSpan, bag)
				}
			}
		}

		ctx, span = tracer.Start(ctx, fmt.Sprintf("ws.%s", message.MessageType),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.MessagingOperationName("process"),
				semconv.MessagingSystemKey.String("websocket"),
				attribute.String("messaging.message.type", message.MessageType),
				attribute.String("component", "websocket-handler"),
				attribute.String("trace.source", "websocket"),
				semconv.MessagingMessageBodySize(len(data)),
			),
		)
	} else {
		ctx, span = tracer.Start(ctx, fmt.Sprintf("ws.%s", message.MessageType),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.MessagingOperationName("process"),
				semconv.MessagingSystemKey.String("websocket"),
				attribute.String("messaging.message.type", message.MessageType),
				attribute.String("component", "websocket-handler"),
				attribute.String("trace.source", "backend"),
				semconv.MessagingMessageBodySize(len(data)),
			),
		)
	}

	defer func() {
		latencyMs := float64(time.Since(start).Milliseconds())
		err = telemetry.RecordRequestLatency(ctx, latencyMs, message.MessageType, messageStatus)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to record latency metric", slog.Any("error", err))
		}
		span.End()
	}()

	err = telemetry.IncrementMessageReceived(ctx, message.MessageType)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to increment message type metric", slog.Any("error", err))
	}

	telemetry.AddMessagingAttributes(ctx, message.MessageType, "process", len(data))
	telemetry.AddWebSocketMetrics(ctx, message.MessageType, len(data), 1, "unicast")

	span.SetAttributes(attribute.String("message_type", message.MessageType))

	if client.playerID != uuid.Nil {
		ctx = telemetry.AddPlayerToBaggage(ctx, client.playerID)
		span.SetAttributes(attribute.String("game.player_id", client.playerID.String()))
	}

	s.logger.DebugContext(ctx, "handling message", slog.String("message_type", message.MessageType))

	ctx = context.WithValue(ctx, "raw_json", data)

	err = s.handlerRegistry.Handle(message.MessageType, ctx, client, s)
	if err != nil {
		var handlerNotFoundErr ErrHandlerNotFound
		if errors.As(err, &handlerNotFoundErr) {
			messageStatus = "fail_matching_handler"
			telemetry.RecordHandlerError(ctx, message.MessageType, err, map[string]interface{}{
				"error_type":   "handler_not_found",
				"message_type": message.MessageType,
			})
		} else {
			var validationErr ErrValidation
			if errors.As(err, &validationErr) {
				messageStatus = "fail_validate"

				telemetry.RecordValidationError(ctx, message.MessageType, validationErr.Err.Error(), "")
				telemetry.RecordValidationErrorMetric(ctx)
				translatedError := translateValidationError(ctx, validationErr.Err.Error())
				webSocketErr := s.updateClientAboutErr(ctx, client.playerID, translatedError)
				if webSocketErr != nil {
					return ctx, errors.Join(err, webSocketErr)
				}
			} else {
				messageStatus = "fail_handler"
				telemetry.RecordHandlerError(ctx, message.MessageType, err, map[string]interface{}{
					"error_type":   "general_handler_error",
					"message_type": message.MessageType,
				})
			}
		}
		return ctx, fmt.Errorf("error in handler function: %w", err)
	}

	s.logger.DebugContext(ctx, "finished handling request")
	span.SetStatus(codes.Ok, "handle_message_successful")
	return ctx, nil
}

func translateValidationError(ctx context.Context, errMsg string) string {
	switch {
	case strings.Contains(errMsg, "player_nickname is required"):
		return i18n.T(ctx, "validation.player_nickname_required")
	case strings.Contains(errMsg, "room_code is required"):
		return i18n.T(ctx, "validation.room_code_required")
	case strings.Contains(errMsg, "game_name is required"):
		return i18n.T(ctx, "validation.game_name_required")
	case strings.Contains(errMsg, "answer is required"):
		return i18n.T(ctx, "validation.answer_required")
	case strings.Contains(errMsg, "player_nickname_to_kick is required"):
		return i18n.T(ctx, "validation.player_nickname_to_kick_required")
	case strings.Contains(errMsg, "player nickname is required"):
		return i18n.T(ctx, "validation.player_nickname_required_voting")
	default:
		return errMsg
	}
}

func (s *Subscriber) newStateDependencies() (*statemachine.StateDependencies, error) {
	timings := s.config.Timings

	return &statemachine.StateDependencies{
		RoundService:  s.roundService,
		ClientUpdater: s,
		Transitioner:  s,
		Logger:        s.logger,
		Timings: statemachine.Timings{
			ShowQuestionScreenFor: timings.ShowQuestionScreenFor,
			ShowVotingScreenFor:   timings.ShowVotingScreenFor,
			ShowRevealScreenFor:   timings.ShowRevealScreenFor,
			ShowScoreScreenFor:    timings.ShowScoreScreenFor,
			ShowWinnerScreenFor:   timings.ShowWinnerScreenFor,
		},
		Scoring: service.Scoring{
			GuessedFibber:      s.config.Scoring.GuessFibber,
			FibberEvadeCapture: s.config.Scoring.FibberEvadeCapture,
		},
	}, nil
}

func (s *Subscriber) StartStateMachine(ctx context.Context, gameStateID uuid.UUID, state statemachine.State) {
	s.stateMachines.Start(ctx, gameStateID, state)
}

func (s *Subscriber) CancelAllStateMachines(ctx context.Context) {
	s.stateMachines.CancelAll(ctx)
}

func (s *Subscriber) WaitForStateMachines(ctx context.Context, timeout time.Duration) bool {
	return s.stateMachines.Wait(ctx, timeout)
}
