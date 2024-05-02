package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/robertjshirts/trademailer/consumer"
	"github.com/robertjshirts/trademailer/dal"
)

func main() {
	dbConfig := ReadDatabaseConfig("config/database.config")
	db, err := dal.Init(dbConfig["host"], dbConfig["port"], dbConfig["user"], dbConfig["password"], dbConfig["protocol"], dbConfig["database"])
	if err != nil {
		log.Panicf("Error initializing database: %v", err)
	}
	defer db.Close()

	mailerConfig := ReadMailerConfig("config/mailer.config")

	kafkaConfig := ReadSaramaConfig("config/kafka.config")
	brokers := strings.Split(kafkaConfig["brokers"], ",")
	topics := []string{kafkaConfig["offerTopic"], kafkaConfig["userTopic"]}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Should add mailerConfig user and pass if using a real mailer
	consumer, err := consumer.Init(db, brokers, kafkaConfig["group"], topics, mailerConfig["host"], mailerConfig["port"])
	if err != nil {
		log.Panicf("Error initializing consumer: %v", err)
	}
	defer consumer.Close()

	fmt.Println("Consumer initialized")
	consumer.Consume(ctx)

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
	cancel()
}
