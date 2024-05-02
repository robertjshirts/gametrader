package dal

type GameCondition string

const (
	Mint GameCondition = "mint"
	Good GameCondition = "good"
	Fair GameCondition = "fair"
	Poor GameCondition = "poor"
)

type StatusCondition string

const (
	Pending   StatusCondition = "pending"
	Accepted  StatusCondition = "accepted"
	Rejected  StatusCondition = "rejected"
	Cancelled StatusCondition = "cancelled"
)

type User struct {
	UserId   *int    `json:"userId"`
	Email    *string `json:"email"`
	Name     *string `json:"name"`
	Address  *string `json:"address"`
	Password *string `json:"password"`
}

type Game struct {
	GameId    *int           `json:"gameId"`
	UserId    *int           `json:"userId"`
	Name      *string        `json:"name"`
	Publisher *string        `json:"publisher"`
	Year      *int           `json:"year"`
	System    *string        `json:"system"`
	Condition *GameCondition `json:"condition"`
	Owners    *int           `json:"owners,omitempty"`
}

// some code
type Offer struct {
	OfferId         *int            `json:"offerId"`
	OffererUserId   *int            `json:"offererUserId"`
	OffererGameId   *int            `json:"offererGameId"`
	RecipientUserId *int            `json:"recipientUserId"`
	RecipientGameId *int            `json:"recipientGameId"`
	Status          StatusCondition `json:"status"`
}
