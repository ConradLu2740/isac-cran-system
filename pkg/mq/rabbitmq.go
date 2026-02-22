package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]interface{}
}

func NewMessageQueue(url string) (*MessageQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &MessageQueue{conn: conn, channel: channel}, nil
}

func (mq *MessageQueue) DeclareQueue(config QueueConfig) (amqp.Queue, error) {
	return mq.channel.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)
}

func (mq *MessageQueue) Publish(ctx context.Context, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = mq.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Message published to queue %s", queueName)
	return nil
}

func (mq *MessageQueue) Consume(ctx context.Context, queueName string, handler func([]byte) error) error {
	msgs, err := mq.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if err := handler(msg.Body); err != nil {
					log.Printf("Message handling error: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	log.Printf("Consumer started for queue %s", queueName)
	return nil
}

func (mq *MessageQueue) Close() error {
	if mq.channel != nil {
		mq.channel.Close()
	}
	if mq.conn != nil {
		mq.conn.Close()
	}
	return nil
}

type AlgorithmTask struct {
	TaskID     string                 `json:"task_id"`
	TaskType   string                 `json:"task_type"`
	Params     map[string]interface{} `json:"params"`
	CreatedAt  int64                  `json:"created_at"`
	Priority   int                    `json:"priority"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
}

type SensorDataMessage struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
	Quality   float64 `json:"quality"`
}

type NotificationMessage struct {
	Type      string                 `json:"type"`
	Level     string                 `json:"level"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

const (
	QueueAlgorithmTask   = "algorithm.task"
	QueueAlgorithmResult = "algorithm.result"
	QueueSensorData      = "sensor.data"
	QueueNotification    = "notification.alert"
)

func (mq *MessageQueue) SetupQueues() error {
	queues := []QueueConfig{
		{Name: QueueAlgorithmTask, Durable: true},
		{Name: QueueAlgorithmResult, Durable: true},
		{Name: QueueSensorData, Durable: true},
		{Name: QueueNotification, Durable: true},
	}

	for _, config := range queues {
		_, err := mq.DeclareQueue(config)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", config.Name, err)
		}
	}

	return nil
}
