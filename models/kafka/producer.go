package kafka

import (
	"context"
	"fmt"
	"rest/myerrors"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	Prod *kafka.Writer
	Cons *kafka.Reader
}

func NewKafka() *Kafka {
	// w := kafka.NewWriter(kafka.WriterConfig{
	// 	Brokers: []string{broker1Address, broker2Address, broker3Address},
	// 	Topic:   topic,
	// })
	return &Kafka{
		Prod: &kafka.Writer{
			Addr:     kafka.TCP("127.0.0.1:9092"),
			Balancer: &kafka.LeastBytes{},
			Topic:    "hash-log",
		},
		Cons: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "hash-log",
			GroupID: "my-group",
		}),
	}
}

func (k *Kafka) Write(ctx context.Context) error {
	// initialize a counter
	key, ok := ctx.Value("key").(string)
	if !ok {
		return myerrors.ErrCtxValue
	}
	val, ok := ctx.Value("val").(string)
	if !ok {
		return myerrors.ErrCtxValue
	}

	err := k.Prod.WriteMessages(ctx, kafka.Message{
		Key: []byte(key),
		// create an arbitrary message payload for the value
		Value: []byte(val),
	})
	if err != nil {
		return err
	}
	return nil
}

func (k *Kafka) Read(ctx context.Context) error {
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages

	for {
		// the `ReadMessage` method blocks until we receive the next event
		msg, err := k.Cons.ReadMessage(ctx)
		if err != nil {
			return err
		}
		// after receiving the message, log its value
		fmt.Println("received: ", string(msg.Value))
	}
}
