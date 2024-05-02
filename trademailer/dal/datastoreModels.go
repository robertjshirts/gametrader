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

type Game struct {
	GameId     int
	Name       string
	System     string
	Conidition GameCondition
}

type User struct {
	Email    string
	Name     string
	Password string
}

type Offer struct {
	OffererUserId   int
	RecipientUserId int
}
