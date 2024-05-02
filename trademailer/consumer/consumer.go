package consumer

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"time"

	"github.com/IBM/sarama"

	"github.com/robertjshirts/trademailer/dal"
)

var smtpServer = "smtp.gmail.com"
var smtpPort = "587"

type Datastore interface {
	GetOfferDetails(offerId int) (*dal.Offer, error)
	GetUserDetails(userId int) (*dal.User, error)
}

type KafkaConsumer struct {
	db            Datastore
	ConsumerGroup sarama.ConsumerGroup
	Topics        []string
}

func Init(db Datastore, brokers []string, group string, topics []string, server string, port string) (*KafkaConsumer, error) {
	smtpServer = server
	smtpPort = port

	config := sarama.NewConfig()
	config.Version = sarama.V3_3_0_0
	config.Consumer.Return.Errors = true

	var consumerGroup sarama.ConsumerGroup
	var err error

	connected := false
	for attempts := 0; attempts < 5; attempts++ {
		consumerGroup, err = sarama.NewConsumerGroup(brokers, group, config)
		if err != nil {
			fmt.Println("Failed to connect to the Kafka cluster. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		} else {
			connected = true
			break
		}
	}

	if !connected {
		return nil, fmt.Errorf("failed to connect to the Kafka cluster after 5 attempts")
	}

	fmt.Printf("Connected to the Kafka cluster at %v \n with topics %v\n", brokers, topics)

	fmt.Println("Waiting 5 seconds for producers to initialize topics...")
	time.Sleep(6 * time.Second)

	if db == nil {
		return nil, fmt.Errorf("no datastore provided")
	}

	return &KafkaConsumer{
		db:            db,
		ConsumerGroup: consumerGroup,
		Topics:        topics,
	}, nil
}

func (ks *KafkaConsumer) Close() error {
	return ks.ConsumerGroup.Close()
}

func (kc *KafkaConsumer) Consume(ctx context.Context) {
	for {
		handler := &consumerGroupHandler{
			db: kc.db,
		}
		err := kc.ConsumerGroup.Consume(ctx, kc.Topics, handler)
		if err != nil {
			log.Panicf("Error consuming: %v", err)
		}
	}
}

type consumerGroupHandler struct {
	db Datastore
}

func (consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		topic := message.Topic
		key := string(message.Key)
		value := string(message.Value)
		fmt.Printf("Message claimed: Topic: %s | Key: %s | Value: %s\n", topic, key, value)

		switch topic {
		case "offer":
			offerId, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("Error converting offer id to int: %v\n", err)
				break
			}

			offer, err := h.db.GetOfferDetails(offerId)
			if err != nil {
				fmt.Printf("Error getting offer details: %v\n", err)
				break
			}

			offerer, err := h.db.GetUserDetails(offer.OffererUserId)
			if err != nil {
				fmt.Printf("Error getting offerer details: %v\n", err)
				break
			}

			recipient, err := h.db.GetUserDetails(offer.RecipientUserId)
			if err != nil {
				fmt.Printf("Error getting recipient details: %v\n", err)
				break
			}

			sendOffererEmail(offerer, recipient, key)
			sendRecipientEmail(recipient, offerer, key)
		case "user":
			userId, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("Error converting user id to int: %v\n", err)
				break
			}

			user, err := h.db.GetUserDetails(userId)
			if err != nil {
				fmt.Printf("Error getting user details: %v\n", err)
				break
			}

			sendUserEmail(user, key)
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func sendOffererEmail(offerer *dal.User, recipient *dal.User, event string) {

	to := []string{recipient.Email}
	auth := smtp.PlainAuth("", offerer.Email, offerer.Password, smtpServer)
	var msg = []byte{}

	switch event {
	case "created":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Successfully Created\r\n"+
			"\r\n"+
			"Congratulations, %s! Your offer to %s was successfully created!",
			offerer.Email, offerer.Name, recipient.Name))
	case "accepted":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Accepted\r\n"+
			"\r\n"+
			"Congratulations, %s! Your offer to %s was accepted!",
			offerer.Email, offerer.Name, recipient.Name))
	case "rejected":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Rejected\r\n"+
			"\r\n"+
			"Sorry, %s. Your offer to %s was rejected.",
			offerer.Email, offerer.Name, recipient.Name))
	case "cancelled":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Cancelled\r\n"+
			"\r\n"+
			"Hey there, %s. Your offer to %s was cancelled.",
			offerer.Email, offerer.Name, recipient.Name))
	}

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, offerer.Email, to, msg)
	if err != nil {
		fmt.Printf("Error sending email: %v\n", err)
	}
}

func sendRecipientEmail(recipient *dal.User, offerer *dal.User, event string) {

	to := []string{recipient.Email}
	auth := smtp.PlainAuth("", recipient.Email, recipient.Password, smtpServer)
	var msg = []byte{}

	switch event {
	case "created":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Received\r\n"+
			"\r\n"+
			"Congratulations, %s! You received an offer from %s.",
			recipient.Email, recipient.Name, offerer.Name))
	case "accepted":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Accepted\r\n"+
			"\r\n"+
			"Congratulations, %s! You accepted an offer from %s.",
			recipient.Email, recipient.Name, offerer.Name))
	case "rejected":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Rejected\r\n"+
			"\r\n"+
			"Hey there, %s. You rejected an offer from %s.",
			recipient.Email, recipient.Name, offerer.Name))
	case "cancelled":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Offer Cancelled\r\n"+
			"\r\n"+
			"Hey there, %s. The offer from %s was cancelled.",
			recipient.Email, recipient.Name, offerer.Name))
	}

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, recipient.Email, to, msg)
	if err != nil {
		fmt.Printf("Error sending email: %v\n", err)
	}
}

func sendUserEmail(user *dal.User, event string) {
	to := []string{user.Email}
	auth := smtp.PlainAuth("", user.Email, user.Password, smtpServer)
	var msg = []byte{}

	switch event {
	case "created":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Gametrader Account Created\r\n"+
			"\r\n"+
			"Welcome, %s! Your account was successfully created.",
			user.Email, user.Name))
	case "updated":
		msg = []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: Gametrader Account Password Updated\r\n"+
			"\r\n"+
			"Hey there, %s. Your account's password was successfully updated.",
			user.Email, user.Name))
	}

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, user.Email, to, msg)
	if err != nil {
		fmt.Printf("Error sending email: %v\n", err)
	}
}
