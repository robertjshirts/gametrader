package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetUser(id UserId) (*UserResponse, error)
	CreateUser(user *PostUser) (*UserResponse, error)
	UpdateUser(id UserId, user *PatchUser) error
	DeleteUser(id UserId) error

	GetGame(id GameId) (*GameResponse, error)
	GetGames(params *GetGamesParams) (*GameSearchResponse, error)
	CreateGame(game *PostGame) (*GameResponse, error)
	UpdateGame(id GameId, game *PatchGame) error
	DeleteGame(id GameId) error

	GetOffer(id OfferId) (*OfferResponse, error)
	GetOffers(params *GetOffersParams) (*OfferSearchResponse, error)
	CreateOffer(offer *PostOffer) (*OfferResponse, error)
	UpdateOffer(id OfferId, offer *PatchOffer) error
	DeleteOffer(id OfferId) error
}

type GameTrader struct {
	service Service
}

func Init(service Service) *GameTrader {
	return &GameTrader{service}
}

//------------------- User -------------------//

func (g *GameTrader) CreateUser(c *gin.Context) {
	var postUserData PostUser
	err := c.BindJSON(&postUserData)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := g.service.CreateUser(&postUserData)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (g *GameTrader) GetUser(c *gin.Context, userId UserId) {
	user, err := g.service.GetUser(userId)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (g *GameTrader) UpdateUser(c *gin.Context, userId UserId) {
	var patchUserData PatchUser
	err := c.BindJSON(&patchUserData)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = g.service.UpdateUser(userId, &patchUserData)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusNoContent)
}

func (g *GameTrader) DeleteUser(c *gin.Context, userId UserId) {
	err := g.service.DeleteUser(userId)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}

//------------------- Game -------------------//

func (g *GameTrader) CreateGame(c *gin.Context) {
	var postGameData PostGame
	err := c.BindJSON(&postGameData)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	game, err := g.service.CreateGame(&postGameData)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, game)
}

func (g *GameTrader) GetGame(c *gin.Context, gameId GameId) {
	game, err := g.service.GetGame(gameId)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, game)
}

func (g *GameTrader) GetGames(c *gin.Context, params GetGamesParams) {
	games, err := g.service.GetGames(&params)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, games)
}

func (g *GameTrader) UpdateGame(c *gin.Context, gameId GameId) {
	var patchGameData PatchGame

	err := c.BindJSON(&patchGameData)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = g.service.UpdateGame(gameId, &patchGameData)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (g *GameTrader) DeleteGame(c *gin.Context, gameId GameId) {
	err := g.service.DeleteGame(gameId)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

//------------------- Offer -------------------//

func (g *GameTrader) CreateOffer(c *gin.Context) {
	var postOfferData PostOffer
	err := c.BindJSON(&postOfferData)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	offer, err := g.service.CreateOffer(&postOfferData)
	if err != nil {
		// Degugging purposes
		c.Error(err)
		c.Status(http.StatusBadRequest)
	}

	c.JSON(http.StatusCreated, offer)
}

func (g *GameTrader) GetOffer(c *gin.Context, offerId OfferId) {
	offer, err := g.service.GetOffer(offerId)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (g *GameTrader) GetOffers(c *gin.Context, params GetOffersParams) {
	offers, err := g.service.GetOffers(&params)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, offers)
}

func (g *GameTrader) UpdateOffer(c *gin.Context, offerId OfferId) {
	var patchOfferData PatchOffer
	err := c.BindJSON(&patchOfferData)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = g.service.UpdateOffer(offerId, &patchOfferData)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (g *GameTrader) DeleteOffer(c *gin.Context, offerId OfferId) {
	err := g.service.DeleteOffer(offerId)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}
