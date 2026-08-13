package kafka

import (
	"context"
	"encoding/json"

	"feature-flags/internal/flags"
	"github.com/segmentio/kafka-go"
)

type Producer struct{ writer *kafka.Writer }

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{writer: &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: topic, Balancer: &kafka.Hash{}}}
}

func (p *Producer) Publish(ctx context.Context, event flags.Event) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(event.Flag.Name), Value: value})
}

func (p *Producer) Close() error { return p.writer.Close() }
