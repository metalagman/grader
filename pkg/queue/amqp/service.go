package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
	"grader/pkg/queue"
	"grader/pkg/workerpool"
	"reflect"
	"runtime"
)

var _ queue.Topic = (*Topic)(nil)
var _ queue.Queue = (*Service)(nil)

type Service struct {
	conn *rabbitmq.Connection
	ch   *rabbitmq.Channel
}

func (s *Service) Channel() *rabbitmq.Channel {
	return s.ch
}

func New(cfg Config) (*Service, error) {
	conn, err := rabbitmq.Dial(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	sendCh, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %w", err)
	}

	s := &Service{
		conn: conn,
		ch:   sendCh,
	}

	return s, nil
}

func (s *Service) Stop() {
	_ = s.ch.Close()
	_ = s.conn.Close()
}

func (s *Service) Topic(topic string) (queue.Topic, error) {
	queueName := topic + "-queue"

	if err := s.ch.ExchangeDeclare(
		topic,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("exchange declare: %w", err)
	}

	_, err := s.ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("queue declare: %w", err)
	}

	if err := s.ch.QueueBind(
		queueName,
		"",
		topic,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("queue bind: %w", err)
	}

	return &Topic{
		exchangeName: topic,
		queueName:    queueName,
		channel:      s.Channel(),
	}, nil
}

type Topic struct {
	exchangeName string
	queueName    string
	channel      *rabbitmq.Channel
}

// Publish message in the topic
func (t *Topic) Publish(message interface{}) error {
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	if err := t.channel.Publish(t.exchangeName, "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	}); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

// Consume message from topic
func (t *Topic) Consume(targetMsg interface{}, consumer queue.ConsumerFunc) error {
	v := reflect.ValueOf(targetMsg)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("want msgType as pointer to struct")
	}

	v = v.Elem() // dereference pointer
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("want msgType as pointer to struct")
	}

	msgType := v.Type()

	messages, err := t.channel.Consume(
		t.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	pool := workerpool.New()
	pool.Start(runtime.GOMAXPROCS(0) * 2)

	for amqpMsg := range messages {
		pool.Run(processMessage(amqpMsg, msgType, consumer))
	}

	pool.Stop()

	return nil
}

func processMessage(amqpMsg amqp.Delivery, msgType reflect.Type, consumer queue.ConsumerFunc) workerpool.Job {
	return func(ctx context.Context) error {
		v := reflect.New(msgType)

		if err := json.Unmarshal(amqpMsg.Body, &v); err != nil {
			if err := amqpMsg.Ack(false); err != nil {
				return fmt.Errorf("amqp acknowledge: %w", err)
			}
			return fmt.Errorf("json decode: %w", err)
		}

		if err := consumer(v); err != nil {
			if err := amqpMsg.Reject(true); err != nil {
				return fmt.Errorf("amqp reject: %w", err)
			}
			return fmt.Errorf("consumer: %w", err)
		}

		if err := amqpMsg.Ack(false); err != nil {
			return fmt.Errorf("amqp acknowledge: %w", err)
		}

		return nil
	}
}
