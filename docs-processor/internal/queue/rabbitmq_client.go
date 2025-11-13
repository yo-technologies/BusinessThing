package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"docs-processor/internal/domain"
	"docs-processor/internal/logger"

	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQClient(url, queueName string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQClient{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *RabbitMQClient) PublishJob(ctx context.Context, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "queue.RabbitMQClient.PublishJob")
	defer span.Finish()

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",
		c.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish job: %w", err)
	}

	return nil
}

type JobHandler func(context.Context, *domain.ProcessingJob) error

func (c *RabbitMQClient) ConsumeJobs(ctx context.Context, consumerTag string, prefetchCount int, handler JobHandler) error {
	err := c.channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		c.queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	logger.Info(ctx, "Worker started, waiting for jobs", "queue", c.queueName, "consumer", consumerTag)

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "Worker shutting down")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				logger.Warn(ctx, "Channel closed")
				return nil
			}

			c.handleMessage(ctx, msg, handler)
		}
	}
}

func (c *RabbitMQClient) handleMessage(ctx context.Context, msg amqp.Delivery, handler JobHandler) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "queue.RabbitMQClient.handleMessage")
	defer span.Finish()

	var job domain.ProcessingJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		logger.Error(ctx, "Failed to unmarshal job", "error", err)
		msg.Nack(false, false)
		return
	}

	logger.Info(ctx, "Processing job", "document_id", job.DocumentID, "retry_count", job.RetryCount)

	if err := handler(ctx, &job); err != nil {
		logger.Error(ctx, "Job processing failed", "error", err, "document_id", job.DocumentID)

		if job.CanRetry() {
			job.IncrementRetry()
			if requeueErr := c.PublishJob(ctx, &job); requeueErr != nil {
				logger.Error(ctx, "Failed to requeue job", "error", requeueErr)
			} else {
				logger.Info(ctx, "Job requeued", "document_id", job.DocumentID, "retry_count", job.RetryCount)
			}
		} else {
			logger.Error(ctx, "Job exceeded max retries", "document_id", job.DocumentID)
		}

		msg.Nack(false, false)
		return
	}

	logger.Info(ctx, "Job completed successfully", "document_id", job.DocumentID)
	msg.Ack(false)
}
