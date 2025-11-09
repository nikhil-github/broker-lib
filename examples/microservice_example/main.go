package main

import (
	"context"
	"fmt"
	"log"
	"time"

	brokerlib "github.com/nikhil-github/broker-lib"
	_ "github.com/nikhil-github/broker-lib/nats"
)

// Example microservice with multiple consumers and publishers
func main() {
	// Single broker connection for the entire microservice
	broker, err := brokerlib.NewBroker("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to create broker: %v", err)
	}
	defer broker.Close()

	// ===== SETUP MULTIPLE CONSUMERS WITH DIFFERENT CONFIGS =====

	// Consumer 1: Bet Request Processor
	betRequestConfig := brokerlib.DefaultConsumerConfig("bet-request-processor")
	betRequestConfig.DeliverPolicy = brokerlib.DeliverLast // Start from last message
	betRequestConfig.MaxDeliveries = 3                     // Max 3 retries
	betRequestConfig.AckWait = 10 * time.Second

	betRequestConsumer, err := broker.NewConsumer("bet-requests", "bet-request-processor", betRequestConfig)
	if err != nil {
		log.Fatalf("Failed to create bet-request consumer: %v", err)
	}

	// Consumer 2: Payment Processor
	paymentConfig := brokerlib.DefaultConsumerConfig("payment-processor")
	paymentConfig.DeliverPolicy = brokerlib.DeliverAll // Start from beginning
	paymentConfig.MaxDeliveries = 5                    // Max 5 retries
	paymentConfig.AckWait = 30 * time.Second

	paymentConsumer, err := broker.NewConsumer("payments", "payment-processor", paymentConfig)
	if err != nil {
		log.Fatalf("Failed to create payment consumer: %v", err)
	}

	// Consumer 3: Order Processor
	orderConfig := brokerlib.DefaultConsumerConfig("order-processor")
	orderConfig.DeliverPolicy = brokerlib.DeliverNew // Only new messages
	orderConfig.AckWait = 60 * time.Second

	orderConsumer, err := broker.NewConsumer("orders", "order-processor", orderConfig)
	if err != nil {
		log.Fatalf("Failed to create order consumer: %v", err)
	}

	// ===== SETUP MULTIPLE PUBLISHERS FOR DIFFERENT STREAMS =====

	payoutPublisher, _ := broker.NewPublisher("payouts")
	ratesPublisher, _ := broker.NewPublisher("rates")

	// ===== EXAMPLE USAGE =====

	ctx := context.Background()

	// Process bet requests
	go func() {
		for {
			messages, err := betRequestConsumer.PullBatch(ctx, 10, 5*time.Second)
			if err != nil {
				log.Printf("Bet request consumer error: %v", err)
				continue
			}
			for _, msg := range messages {
				fmt.Printf("[BetRequest] Processing: %s\n", string(msg.Data))
				// Process bet request...
				msg.Ack()
			}
		}
	}()

	// Process payments
	go func() {
		for {
			messages, err := paymentConsumer.PullBatch(ctx, 5, 10*time.Second)
			if err != nil {
				log.Printf("Payment consumer error: %v", err)
				continue
			}
			for _, msg := range messages {
				fmt.Printf("[Payment] Processing: %s\n", string(msg.Data))
				// Process payment...
				msg.Ack()

				// Publish payout after successful payment
				payoutPublisher.Publish([]byte(fmt.Sprintf("Payout for: %s", string(msg.Data))))
			}
		}
	}()

	// Process orders
	go func() {
		for {
			messages, err := orderConsumer.PullBatch(ctx, 20, 3*time.Second)
			if err != nil {
				log.Printf("Order consumer error: %v", err)
				continue
			}
			for _, msg := range messages {
				fmt.Printf("[Order] Processing: %s\n", string(msg.Data))
				// Process order...
				msg.Ack()

				// Publish rate update
				ratesPublisher.Publish([]byte(fmt.Sprintf("Rate update for: %s", string(msg.Data))))
			}
		}
	}()

	// Keep running
	fmt.Println("Microservice running with multiple consumers and publishers...")
	time.Sleep(30 * time.Second)
}
