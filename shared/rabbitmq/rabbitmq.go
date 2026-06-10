package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/aliamerj/wardu/shared/env"
	amqp "github.com/rabbitmq/amqp091-go"
	zlog "github.com/rs/zerolog/log"
)

const (
	JobsExchange = "wardu.jobs"
	JobsDLX      = "wardu.jobs.dlx"

	JobsQueue   = "wardu.jobs"
	FailedQueue = "wardu.jobs.failed"
)

type JobMessage struct {
	JobID    string `json:"job_id"`
	Attempt  int    `json:"attempt"`
	Image    string `json:"image"`
	Priority int64  `json:"priority"`
}

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

var rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

func New() (*RabbitMQ, error) {
	conn, err := amqp.Dial(rabbitMqURI)
	if err != nil {
		return nil, err
	}
	zlog.Info().Msg("rabbitmq connected successfuly")

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	r := &RabbitMQ{
		conn: conn,
		ch:   ch,
	}

	if err := r.setup(); err != nil {
		r.Close()
		return nil, err
	}
	zlog.Info().Msg("rabbitmq setup successfuly")

	return r, nil
}

func (r *RabbitMQ) Close() {
	if r.conn == nil {
		return
	}

	if err := r.conn.Close(); err != nil {
		zlog.Error().Err(err).Msg("failed to close rabbitmq")
	}
}

func (r *RabbitMQ) PublishJob(
	ctx context.Context,
	msg JobMessage,
) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return r.ch.PublishWithContext(
		ctx,
		JobsExchange,
		"job",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Priority:     uint8(msg.Priority),
			Body:         body,
		},
	)
}

func (r *RabbitMQ) setup() error {
	if err := r.ch.ExchangeDeclare(
		JobsExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := r.ch.ExchangeDeclare(
		JobsDLX,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	_, err := r.ch.QueueDeclare(
		JobsQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange": JobsDLX,
		},
	)
	if err != nil {
		return err
	}

	_, err = r.ch.QueueDeclare(
		FailedQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if err := r.ch.QueueBind(
		JobsQueue,
		"job",
		JobsExchange,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := r.ch.QueueBind(
		FailedQueue,
		"failed",
		JobsDLX,
		false,
		nil,
	); err != nil {
		return err
	}
	return nil
}
