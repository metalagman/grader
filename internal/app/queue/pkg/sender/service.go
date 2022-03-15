package sender

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"grader/internal/pkg/model"
	"grader/pkg/logger"
	"grader/pkg/queue"
)

type Sender struct {
	topic  queue.Topic
	client *resty.Client
}

func New(
	q queue.Queue,
	topicName string,
) (*Sender, error) {
	t, err := q.Topic(topicName)
	if err != nil {
		return nil, err
	}

	return &Sender{
		topic:  t,
		client: resty.New(),
	}, nil
}

func (s *Sender) Send() error {
	l := logger.Global()

	err := s.topic.Consume(model.Submission{}, func(ctx context.Context, message interface{}) error {
		val, ok := message.(*model.Submission)
		if !ok {
			return fmt.Errorf("invalid message type, got %T %#v", message, message)
		}

		//req := &runner.Submission{
		//
		//}

		/**
		ContainerImage string           `json:"container_image" validate:"required"`
		PartID         string           `json:"part_id" validate:"required"`
		PostbackURL    string           `json:"postback_url" validate:"required,url"`
		Files          []SubmissionFile `json:"files" validate:"required"`
		*/

		//_, err := s.client.R().
		//	SetContext(ctx).
		//	SetHeader("Content-Type", "application/json").
		//	SetBody(req).
		//	Post()
		//if err != nil {
		//	return fmt.Errorf("result postback: %w", err)
		//}

		l.Debug().Msgf("%#v", val)
		return nil
	})

	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	return err
}
