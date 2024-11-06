package websockets

import (
	"context"
	"fmt"
	"time"
)

type RoundServicer interface {
	SubmitAnswer(ctx context.Context, playerID, answer string, submittedAt time.Time) error
}

func (s *SubmitAnswer) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	err := sub.roundService.SubmitAnswer(ctx, client.playerID, s.Answer, time.Now())
	if err != nil {
		errStr := "failed to submit answer, try again"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	return nil
}
