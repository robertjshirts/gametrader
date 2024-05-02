package services

import (
	"fmt"
	"time"

	//"encoding/json"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"

	"github.com/robertjshirts/gobuster/api"
	"github.com/robertjshirts/gobuster/dal"
)

type Datastore interface {
	GetUser(id int) (*dal.User, error)
	CreateUser(user *dal.User) (*dal.User, error)
	UpdateUser(id int, user *dal.User) error
	DeleteUser(id int) error

	GetGame(id int) (*dal.Game, error)
	GetGames(userId *int, offset *int, limit *int) ([]dal.Game, error)
	CreateGame(game *dal.Game) (*dal.Game, error)
	UpdateGame(id int, game *dal.Game) error
	DeleteGame(id int) error

	ChangeGameUserId(id int, userId int) error

	GetOffer(id int) (*dal.Offer, error)
	GetOffers(offererUserId *int, recipientUserId *int, offset *int, limit *int) ([]dal.Offer, error)
	CreateOffer(offer *dal.Offer) (*dal.Offer, error)
	UpdateOffer(id int, offer *dal.Offer) error
	DeleteOffer(id int) error
}

type Service struct {
	db         Datastore
	producer   sarama.SyncProducer
	offerTopic string
	userTopic  string
}

func Init(db Datastore, brokers []string, offerTopic string, userTopic string) (*Service, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_3_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	// Connect to kafka
	fmt.Printf("Connecting to Kafka at %v\n", brokers)
	var producer sarama.SyncProducer
	var err error

	connected := false
	for attempts := 0; attempts < 5; attempts++ {
		producer, err = sarama.NewSyncProducer(brokers, config)
		if err != nil {
			fmt.Printf("Failed to connect to the Kafka cluster. Retrying in 5 seconds...\n")
			time.Sleep(5 * time.Second)
		} else {
			connected = true
			break
		}
	}

	if !connected {
		return nil, fmt.Errorf("failed to connect to the Kafka cluster after 5 attempts")
	}

	fmt.Println("Connected to the Kafka cluster")

	topics := []string{offerTopic, userTopic}
	for _, topic := range topics {
		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder("init"),
			Value: sarama.StringEncoder("init"),
		})

		if err != nil {
			return nil, err
		}

		fmt.Printf("Sent init message to topic %v\n", topic)
	}

	return &Service{
		db:         db,
		producer:   producer,
		offerTopic: offerTopic,
		userTopic:  userTopic}, nil
}

func (s *Service) Close() error {
	return s.producer.Close()
}

// ---------------- Middleware ----------------//

func (s *Service) Middleware(c *gin.Context) {
	c.Next()
}

// ------------------- User -------------------//

func (s *Service) GetUser(id api.UserId) (*api.UserResponse, error) {
	// Call the db method to get the user
	dalUser, err := s.db.GetUser(id)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	apiUser := api.UserResponse{
		UserId:  *dalUser.UserId,
		Email:   *dalUser.Email,
		Name:    *dalUser.Name,
		Address: *dalUser.Address,
	}

	return &apiUser, nil
}

func (s *Service) CreateUser(user *api.PostUser) (*api.UserResponse, error) {
	// Convert the api model to the dal model (dereference the pointers because its on the createUser method)
	dalUser := dal.User{
		Email:    &user.Email,
		Name:     &user.Name,
		Address:  &user.Address,
		Password: &user.Password,
	}

	// Call the db method to create the user
	createdUser, err := s.db.CreateUser(&dalUser)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	apiUser := api.UserResponse{
		UserId:  *createdUser.UserId,
		Email:   *createdUser.Email,
		Name:    *createdUser.Name,
		Address: *createdUser.Address,
	}

	return &apiUser, nil
}

func (s *Service) UpdateUser(id api.UserId, user *api.PatchUser) error {
	// Convert the api model to the dal model (no need to dereference the pointers because its on the updateUser method)
	dalUser := dal.User{
		Name:     user.Name,
		Address:  user.Address,
		Password: user.Password,
	}

	// Call the db method to update the user
	err := s.db.UpdateUser(id, &dalUser)
	if err != nil {
		return err
	}

	// If the password isn't being updated, return
	if user.Password == nil {
		return nil
	}

	// Send password update to the kafka topic
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: s.userTopic,
		Key:   sarama.StringEncoder("updated"),
		Value: sarama.StringEncoder(fmt.Sprint(id)),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteUser(id api.UserId) error {
	return s.db.DeleteUser(id)
}

// ------------------- Game -------------------//

func (s *Service) GetGame(id api.GameId) (*api.GameResponse, error) {
	// Call the db method to get the game
	game, err := s.db.GetGame(id)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	apiGame := api.GameResponse{
		GameId:    *game.GameId,
		UserId:    "/users/" + fmt.Sprint(*game.UserId),
		Name:      *game.Name,
		Publisher: *game.Publisher,
		Year:      *game.Year,
		System:    *game.System,
		Condition: api.GameConditionEnum(*game.Condition),
		Owners:    game.Owners,
	}

	return &apiGame, nil
}

func (s *Service) GetGames(params *api.GetGamesParams) (*api.GameSearchResponse, error) {
	// Parse search params
	userId := params.UserId
	offset := params.Offset
	limit := params.Limit

	// Call the db method to get the games
	dalGames, err := s.db.GetGames(userId, offset, limit)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	var apiGames api.GameSearchResponse
	for _, game := range dalGames {
		apiGame := api.GameResponse{
			GameId:    *game.GameId,
			UserId:    "/users/" + fmt.Sprint(*game.UserId),
			Name:      *game.Name,
			Publisher: *game.Publisher,
			Year:      *game.Year,
			System:    *game.System,
			Condition: api.GameConditionEnum(*game.Condition),
			Owners:    game.Owners,
		}
		apiGames = append(apiGames, apiGame)
	}

	return &apiGames, nil
}

func (s *Service) CreateGame(game *api.PostGame) (*api.GameResponse, error) {
	// Convert the api model to the dal model
	dalGame := dal.Game{
		UserId:    &game.UserId,
		Name:      &game.Name,
		Publisher: &game.Publisher,
		Year:      &game.Year,
		System:    &game.System,
		Condition: s.convertCondition(&game.Condition),
		Owners:    game.Owners,
	}

	// Call the db method to create the game
	createdGame, err := s.db.CreateGame(&dalGame)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	apiGame := api.GameResponse{
		GameId:    *createdGame.GameId,
		UserId:    "/users/" + fmt.Sprint(*createdGame.UserId),
		Name:      *createdGame.Name,
		Publisher: *createdGame.Publisher,
		Year:      *createdGame.Year,
		System:    *createdGame.System,
		Condition: api.GameConditionEnum(*createdGame.Condition),
		Owners:    createdGame.Owners,
	}

	return &apiGame, nil
}

func (s *Service) UpdateGame(id api.GameId, game *api.PatchGame) error {
	// Convert the api model to the dal model
	dalGame := dal.Game{
		Name:      game.Name,
		Publisher: game.Publisher,
		Year:      game.Year,
		System:    game.System,
		Condition: s.convertCondition(game.Condition),
		Owners:    game.Owners,
	}

	// Call the db method to update the game
	err := s.db.UpdateGame(id, &dalGame)
	return err
}

func (s *Service) DeleteGame(id api.GameId) error {
	// Call the db method to delete the game
	return s.db.DeleteGame(id)
}

// ------------------- Offers -------------------//

func (s *Service) GetOffer(id api.OfferId) (*api.OfferResponse, error) {
	// Call the db method to get the offer
	offer, err := s.db.GetOffer(id)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	apiOffer := api.OfferResponse{
		OfferId:         *offer.OfferId,
		OffererUserId:   "/users/" + fmt.Sprint(*offer.OffererUserId),
		OffererGameId:   "/games/" + fmt.Sprint(*offer.OffererGameId),
		RecipientUserId: "/users/" + fmt.Sprint(*offer.RecipientUserId),
		RecipientGameId: "/games/" + fmt.Sprint(*offer.RecipientGameId),
		Status:          api.OfferStatusEnum(offer.Status),
	}

	return &apiOffer, nil
}

func (s *Service) GetOffers(params *api.GetOffersParams) (*api.OfferSearchResponse, error) {
	// Parse search params
	offererUserId := params.OffererUserId
	recipientUserId := params.RecipientUserId
	offset := params.Offset
	limit := params.Limit

	// Call the db method to get the offers
	dalOffers, err := s.db.GetOffers(offererUserId, recipientUserId, offset, limit)
	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	var apiOffers api.OfferSearchResponse
	for _, offer := range dalOffers {
		apiOffer := api.OfferResponse{
			OfferId:         *offer.OfferId,
			OffererUserId:   "/users/" + fmt.Sprint(*offer.OffererUserId),
			OffererGameId:   "/games/" + fmt.Sprint(*offer.OffererGameId),
			RecipientUserId: "/users/" + fmt.Sprint(*offer.RecipientUserId),
			RecipientGameId: "/games/" + fmt.Sprint(*offer.RecipientGameId),
			Status:          api.OfferStatusEnum(offer.Status),
		}
		apiOffers = append(apiOffers, apiOffer)
	}

	return &apiOffers, nil
}

func (s *Service) CreateOffer(offer *api.PostOffer) (*api.OfferResponse, error) {
	// Convert the api model to the dal model
	dalOffer := dal.Offer{
		OffererUserId:   &offer.OffererUserId,
		OffererGameId:   &offer.OffererGameId,
		RecipientUserId: &offer.RecipientUserId,
		RecipientGameId: &offer.RecipientGameId,
		Status:          dal.Pending,
	}

	// Call the db method to create the offer
	createdOffer, err := s.db.CreateOffer(&dalOffer)
	if err != nil {
		return nil, err
	}

	// Verify the offer
	err = s.validateOffer(*createdOffer.OfferId)
	if err != nil {
		return nil, err
	}

	// Send the offer to the kafka topic
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: s.offerTopic,
		Key:   sarama.StringEncoder("created"),
		Value: sarama.StringEncoder(fmt.Sprint(*createdOffer.OfferId)),
	})

	if err != nil {
		return nil, err
	}

	// Convert the dal model to the api model
	apiOffer := api.OfferResponse{
		OfferId:         *createdOffer.OfferId,
		OffererUserId:   "/users/" + fmt.Sprint(*createdOffer.OffererUserId),
		OffererGameId:   "/games/" + fmt.Sprint(*createdOffer.OffererGameId),
		RecipientUserId: "/users/" + fmt.Sprint(*createdOffer.RecipientUserId),
		RecipientGameId: "/games/" + fmt.Sprint(*createdOffer.RecipientGameId),
		Status:          api.OfferStatusEnum(createdOffer.Status),
	}

	return &apiOffer, nil
}

func (s *Service) UpdateOffer(id api.OfferId, offer *api.PatchOffer) error {
	// Convert the api model to the dal model
	dalOffer := dal.Offer{
		Status: s.convertStatus(offer),
	}

	// Verify the offer
	err := s.validateOffer(int(id))
	if err != nil {
		return err
	}

	// Call the db method to update the offer
	err = s.db.UpdateOffer(id, &dalOffer)
	if err != nil {
		return err
	}

	// Update the game owners if the offer was accepted
	if dalOffer.Status == dal.Accepted {
		err = s.executeOffer(int(id))
		if err != nil {
			return err
		}
	}

	// Send the offer to the kafka topic
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: s.offerTopic,
		Key:   sarama.StringEncoder(dalOffer.Status),
		Value: sarama.StringEncoder(fmt.Sprint(id)),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteOffer(id api.OfferId) error {
	// Call the db method to delete the offer
	return s.db.DeleteOffer(id)
}

// ------------------- Helpers -------------------//

func (s *Service) validateOffer(offerId int) error {
	// Retrieve the offer
	offer, err := s.db.GetOffer(offerId)
	if err != nil {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return err
	}

	// Check if the offerer and recipient are different
	if *offer.OffererUserId == *offer.RecipientUserId {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return fmt.Errorf("offerer and recipient cannot be the same user")
	}

	// Check if the offerer and recipient games are different
	if *offer.OffererGameId == *offer.RecipientGameId {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return fmt.Errorf("offerer and recipient games cannot be the same game")
	}

	// Check if the offerer and recipient games are owned by the correct users
	offererGame, err := s.db.GetGame(*offer.OffererGameId)
	if err != nil {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return err
	}
	if *offererGame.UserId != *offer.OffererUserId {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return fmt.Errorf("offerer does not own the offerer game")
	}

	recipientGame, err := s.db.GetGame(*offer.RecipientGameId)
	if err != nil {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return err
	}
	if *recipientGame.UserId != *offer.RecipientUserId {
		s.db.UpdateOffer(offerId, &dal.Offer{Status: dal.Rejected})
		return fmt.Errorf("recipient does not own the recipient game")
	}

	return nil
}

func (s *Service) executeOffer(offerId int) error {
	// Get the offer
	offer, err := s.db.GetOffer(offerId)
	if err != nil {
		return err
	}

	// Verify offer status
	if offer.Status != dal.Accepted {
		return fmt.Errorf("offer status is not accepted, cannot execute trade")
	}

	// Change Offerer's game to Recipient's user
	err = s.db.ChangeGameUserId(*offer.OffererGameId, *offer.RecipientUserId)
	if err != nil {
		return err
	}

	// Change Recipient's game to Offerer's user
	err = s.db.ChangeGameUserId(*offer.RecipientGameId, *offer.OffererUserId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) convertCondition(condition *api.GameConditionEnum) *dal.GameCondition {
	if condition == nil {
		return nil
	}
	converted := dal.GameCondition(*condition)
	return &converted
}

func (s *Service) convertStatus(status *api.OfferStatusEnum) dal.StatusCondition {
	converted := dal.StatusCondition(*status)
	return converted
}
