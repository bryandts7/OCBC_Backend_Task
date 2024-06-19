package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

func main() {
	// Connect to Kafka
	connectKafka()

	// Connect to Redis
	connectRedis()

	// Connect to Jaeger
	connectJaeger()
}

func connectKafka() {
	brokers := []string{"localhost:29092"} // Kafka broker address
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatal("Error creating Kafka producer: ", err)
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: "test_topic", // Kafka topic
		Value: sarama.StringEncoder("Hello Kafka"),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatal("Error sending message to Kafka: ", err)
	}

	fmt.Printf("Message sent to Kafka partition %d with offset %d\n", partition, offset)
}

func connectRedis() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis address
	})

	err := client.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatal("Error setting value in Redis: ", err)
	}

	val, err := client.Get(ctx, "key").Result()
	if err != nil {
		log.Fatal("Error getting value from Redis: ", err)
	}

	fmt.Printf("Redis key: %s\n", val)
}

func connectJaeger() {
	cfg := config.Configuration{
		ServiceName: "my-service",
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: "localhost:6831", // Jaeger agent address
		},
	}

	tracer, closer, err := cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
		config.Metrics(metrics.NullFactory),
	)
	if err != nil {
		log.Fatal("Error initializing Jaeger Tracer: ", err)
	}
	defer closer.Close()

	span := tracer.StartSpan("my-operation")
	defer span.Finish()

	span.SetTag("my-tag", "some-value")
	time.Sleep(2 * time.Second)

	fmt.Println("Tracing information sent to Jaeger")
}
