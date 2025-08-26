package telemetry

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type GameContext struct {
	GameID      *uuid.UUID `json:"game_id,omitempty"`
	PlayerID    *uuid.UUID `json:"player_id,omitempty"`
	GameState   string     `json:"game_state,omitempty"`
	GameStateID *uuid.UUID `json:"game_state_id,omitempty"`
	RoomCode    string     `json:"room_code,omitempty"`
	Round       *int       `json:"round,omitempty"`
}

func AddGameContextToSpan(ctx context.Context, gameCtx GameContext) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	var attrs []attribute.KeyValue

	if gameCtx.GameID != nil {
		attrs = append(attrs, attribute.String("game.id", gameCtx.GameID.String()))
	}
	if gameCtx.PlayerID != nil {
		attrs = append(attrs, attribute.String("game.player_id", gameCtx.PlayerID.String()))
	}
	if gameCtx.GameState != "" {
		attrs = append(attrs, attribute.String("game.state", gameCtx.GameState))
	}
	if gameCtx.GameStateID != nil {
		attrs = append(attrs, attribute.String("game.state_id", gameCtx.GameStateID.String()))
	}
	if gameCtx.RoomCode != "" {
		attrs = append(attrs, attribute.String("game.room_code", gameCtx.RoomCode))
	}
	if gameCtx.Round != nil {
		attrs = append(attrs, attribute.Int("game.round", *gameCtx.Round))
	}

	span.SetAttributes(attrs...)
}

func AddRoomStateAttributes(ctx context.Context, roomState, roomCode string, playerCount int) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("room.state", roomState),
		attribute.String("room.code", roomCode),
		attribute.Int("room.player_count", playerCount),
	)
}

func AddPlayerActionAttributes(ctx context.Context, playerID, action string, isHost, isReady bool) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("player.id", playerID),
		attribute.String("player.action", action),
		attribute.Bool("player.is_host", isHost),
		attribute.Bool("player.is_ready", isReady),
	)
}

func AddGameStateTransition(ctx context.Context, fromState, toState, trigger string, gameStateID *uuid.UUID) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("game.state.from", fromState),
		attribute.String("game.state.to", toState),
		attribute.String("game.state.trigger", trigger),
	}

	if gameStateID != nil {
		attrs = append(attrs, attribute.String("game.state_id", gameStateID.String()))
	}

	span.AddEvent("game.state.transition", trace.WithAttributes(attrs...))
	span.SetAttributes(
		attribute.String("game.state.current", toState),
		attribute.String("game.state.previous", fromState),
	)
}

func AddQuestionAttributes(
	ctx context.Context,
	questionID, questionText, roundType string,
	round int,
	deadline string,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("question.id", questionID),
		attribute.String("question.text", questionText),
		attribute.String("question.round_type", roundType),
		attribute.Int("question.round", round),
		attribute.String("question.deadline", deadline),
	)
}

func AddAnswerAttributes(
	ctx context.Context,
	playerID, answer string,
	isValid bool,
	possibleAnswers []string,
	submissionTime string,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("answer.player_id", playerID),
		attribute.String("answer.text", answer),
		attribute.Bool("answer.is_valid", isValid),
		attribute.String("answer.submission_time", submissionTime),
	}

	if len(possibleAnswers) > 0 {
		attrs = append(attrs, attribute.StringSlice("answer.possible_answers", possibleAnswers))
	}

	span.SetAttributes(attrs...)
}

func AddVotingAttributes(ctx context.Context, voterID, targetPlayerID string, voteCount int, round int) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("vote.voter_id", voterID),
		attribute.String("vote.target_player_id", targetPlayerID),
		attribute.Int("vote.count", voteCount),
		attribute.Int("vote.round", round),
	)
}

func AddScoreAttributes(ctx context.Context, playerID string, score, totalScore int, role string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("score.player_id", playerID),
		attribute.Int("score.round_points", score),
		attribute.Int("score.total_points", totalScore),
		attribute.String("score.player_role", role),
	)
}

func AddGameEndAttributes(ctx context.Context, winnerID, reason string, totalRounds int, gameDuration string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.AddEvent("game.ended", trace.WithAttributes(
		attribute.String("game.winner_id", winnerID),
		attribute.String("game.end_reason", reason),
		attribute.Int("game.total_rounds", totalRounds),
		attribute.String("game.duration", gameDuration),
	))

	span.SetAttributes(
		attribute.String("game.outcome", "completed"),
		attribute.String("game.winner_id", winnerID),
	)
}

func AddPlayerConnectionAttributes(
	ctx context.Context,
	playerID, connectionType string,
	isReconnection bool,
	sessionDuration string,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("connection.player_id", playerID),
		attribute.String("connection.type", connectionType),
		attribute.Bool("connection.is_reconnection", isReconnection),
		attribute.String("connection.session_duration", sessionDuration),
	)
}

func AddTimingAttributes(ctx context.Context, operation string, duration, timeRemaining string, hasExpired bool) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("timing.operation", operation),
		attribute.String("timing.duration", duration),
		attribute.String("timing.time_remaining", timeRemaining),
		attribute.Bool("timing.has_expired", hasExpired),
	)
}

func AddLobbyStatusAttributes(
	ctx context.Context,
	roomCode string,
	totalPlayers, readyPlayers int,
	allReady bool,
	canStart bool,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("lobby.room_code", roomCode),
		attribute.Int("lobby.total_players", totalPlayers),
		attribute.Int("lobby.ready_players", readyPlayers),
		attribute.Bool("lobby.all_ready", allReady),
		attribute.Bool("lobby.can_start", canStart),
		attribute.Float64("lobby.ready_percentage", float64(readyPlayers)/float64(totalPlayers)*100),
	)
}

func AddWebSocketMetrics(
	ctx context.Context,
	messageType string,
	messageSize int,
	clientCount int,
	broadcastType string,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("websocket.message_type", messageType),
		attribute.Int("websocket.message_size_bytes", messageSize),
		attribute.Int("websocket.active_clients", clientCount),
		attribute.String("websocket.broadcast_type", broadcastType),
	)
}

func AddValidationErrorDetails(
	ctx context.Context,
	fieldName, expectedValue, actualValue string,
	validationRule string,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.AddEvent("validation.field_error", trace.WithAttributes(
		attribute.String("validation.field_name", fieldName),
		attribute.String("validation.expected_value", expectedValue),
		attribute.String("validation.actual_value", actualValue),
		attribute.String("validation.rule", validationRule),
	))
}

func AddGameProgressMetrics(
	ctx context.Context,
	gameStateID string,
	playersInGame, playersReady, playersAnswered int,
	phase string,
) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("game.progress.game_state_id", gameStateID),
		attribute.String("game.progress.phase", phase),
		attribute.Int("game.progress.players_in_game", playersInGame),
		attribute.Int("game.progress.players_ready", playersReady),
		attribute.Int("game.progress.players_answered", playersAnswered),
		attribute.Float64("game.progress.completion_rate", float64(playersAnswered)/float64(playersInGame)*100),
	)
}

func AddPlayerToBaggage(ctx context.Context, playerID uuid.UUID) context.Context {
	member, err := baggage.NewMember("player_id", playerID.String())
	if err != nil {
		return ctx
	}

	bag, err := baggage.FromContext(ctx).SetMember(member)
	if err != nil {
		return ctx
	}

	return baggage.ContextWithBaggage(ctx, bag)
}

func AddRoomCodeToBaggage(ctx context.Context, roomCode string) context.Context {
	member, err := baggage.NewMember("room_code", roomCode)
	if err != nil {
		return ctx
	}

	bag, err := baggage.FromContext(ctx).SetMember(member)
	if err != nil {
		return ctx
	}

	return baggage.ContextWithBaggage(ctx, bag)
}

func GetPlayerIDFromBaggage(ctx context.Context) *uuid.UUID {
	bag := baggage.FromContext(ctx)
	member := bag.Member("player_id")
	if member.Value() == "" {
		return nil
	}

	if playerID, err := uuid.FromString(member.Value()); err == nil {
		return &playerID
	}
	return nil
}

func RecordValidationError(ctx context.Context, field, message, value string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.AddEvent("validation.error", trace.WithAttributes(
		attribute.String("validation.field", field),
		attribute.String("validation.message", message),
		attribute.String("validation.value", value),
		semconv.ErrorTypeKey.String("validation"),
	))

	span.SetAttributes(
		attribute.Bool("error", true),
		semconv.ErrorTypeKey.String("validation"),
	)
}

func RecordHandlerError(ctx context.Context, handler string, err error, extra map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		semconv.ErrorTypeKey.String("handler"),
		attribute.String("handler.name", handler),
		attribute.String("exception.message", err.Error()),
		attribute.String("exception.type", "HandlerError"),
	}

	if extra != nil {
		if extraJSON, jsonErr := json.Marshal(extra); jsonErr == nil {
			attrs = append(attrs, attribute.String("error.context", string(extraJSON)))
		}
	}

	span.AddEvent("exception", trace.WithAttributes(attrs...))
	span.SetAttributes(
		attribute.Bool("error", true),
		semconv.ErrorTypeKey.String("handler"),
	)
}

func RecordBusinessLogicError(ctx context.Context, operation string, reason string, gameCtx GameContext) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		semconv.ErrorTypeKey.String("business_logic"),
		attribute.String("operation.name", operation),
		attribute.String("exception.message", reason),
		attribute.String("exception.type", "BusinessLogicError"),
	}

	if gameCtx.GameState != "" {
		attrs = append(attrs, attribute.String("game.state", gameCtx.GameState))
	}
	if gameCtx.RoomCode != "" {
		attrs = append(attrs, attribute.String("game.room_code", gameCtx.RoomCode))
	}
	if gameCtx.PlayerID != nil {
		attrs = append(attrs, attribute.String("game.player_id", gameCtx.PlayerID.String()))
	}

	span.AddEvent("exception", trace.WithAttributes(attrs...))
	span.SetAttributes(
		attribute.Bool("error", true),
		semconv.ErrorTypeKey.String("business_logic"),
	)
}

func AddMessagingAttributes(ctx context.Context, messageType, operation string, payloadSize int) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		semconv.MessagingOperationName(operation),
		semconv.MessagingSystemKey.String("websocket"),
		attribute.String("messaging.protocol", "websocket"),
		attribute.String("messaging.message.type", messageType),
		semconv.MessagingMessageBodySize(payloadSize),
	}

	span.SetAttributes(attrs...)
}

func AddDatabaseAttributes(ctx context.Context, dbName, operation, statement string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		semconv.DBSystemKey.String("postgresql"),
		semconv.DBNamespace(dbName),
		semconv.DBOperationName(operation),
		semconv.DBQueryText(statement),
	}

	span.SetAttributes(attrs...)
}

func StartInternalSpan(
	ctx context.Context,
	tracer trace.Tracer,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	return tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)
}

func StartServerSpan(
	ctx context.Context,
	tracer trace.Tracer,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	return tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...),
	)
}

func StartClientSpan(
	ctx context.Context,
	tracer trace.Tracer,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	return tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
}
