package dal

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Implements all methods in the Datastore interaface
type SQLDatastore struct {
	db *sql.DB
}

// Initializes the database connection to the MySQL database, using the username and password provided,
// using the net protocol and address provided (e.g. "tcp" and "localhost:3306"), and using the database
// name provided.
func Init(user string, pass string, net string, address string, port string, dbName string) (*SQLDatastore, error) {
	// Create the struct
	d := new(SQLDatastore)
	// Capture connection properties.
	cfg := mysql.Config{
		User:   user,
		Passwd: pass,
		Net:    net,
		Addr:   address + ":" + port,
		DBName: dbName,
	}
	// Open the connection
	fmt.Printf("Connecting to database with %s\n", cfg.FormatDSN())
	var err error
	d.db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	connected := false
	for attempts := 0; attempts < 5; attempts++ {
		// Test the connection
		if err := d.db.Ping(); err != nil {
			fmt.Println("Failed to connect to the database. Retrying in 10 seconds...")
			time.Sleep(10 * time.Second)
		} else {
			connected = true
			break
		}
	}

	if !connected {
		return nil, fmt.Errorf("failed to connect to the database after 5 attempts")
	}

	fmt.Println("Connected to the database")

	return d, nil
}

func (d *SQLDatastore) Close() error {
	return d.db.Close()
}

// ------------------- User -------------------//

func (d *SQLDatastore) GetUser(id int) (*User, error) {
	var user User
	err := d.db.QueryRow("SELECT * FROM users WHERE `userId` = ?", id).Scan(&user.UserId, &user.Email, &user.Name, &user.Address)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *SQLDatastore) CreateUser(user *User) (*User, error) {
	result, err := d.db.Exec("INSERT INTO users (`email`, `name`, `address`, `password`) VALUES (?, ?, ?, ?)", user.Email, user.Name, user.Address, user.Password)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	intId := int(id)
	user.UserId = &intId

	return user, nil
}

func (d *SQLDatastore) UpdateUser(id int, user *User) error {
	query := "UPDATE users SET "
	args := []interface{}{}
	updates := []string{}

	if user.Name != nil {
		updates = append(updates, "`name` = ?")
		args = append(args, user.Name)
	}

	if user.Address != nil {
		updates = append(updates, "`address` = ?")
		args = append(args, user.Address)
	}

	if user.Password != nil {
		updates = append(updates, "`password` = ?")
		args = append(args, user.Password)
	}

	if len(updates) == 0 {
		return nil
	}

	query += strings.Join(updates, ", ")
	query += " WHERE userId = ?"
	args = append(args, id)
	_, err := d.db.Exec(query, args...)
	return err
}

func (d *SQLDatastore) DeleteUser(id int) error {
	_, err := d.db.Exec("DELETE FROM users WHERE `userId` = ?", id)
	return err
}

// ------------------- Game -------------------//

func (d *SQLDatastore) GetGame(id int) (*Game, error) {
	var game Game
	err := d.db.QueryRow("SELECT * FROM games WHERE `gameId` = ?", id).Scan(&game.GameId, &game.UserId, &game.Name, &game.Publisher, &game.Year, &game.System, &game.Condition, &game.Owners)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (d *SQLDatastore) GetGames(userId *int, offset *int, limit *int) ([]Game, error) {
	var games []Game
	var rows *sql.Rows
	var err error
	query := "SELECT * FROM games"
	args := []interface{}{}

	if userId != nil {
		query += " WHERE `userId` = ?"
		args = append(args, *userId)
	}
	if limit != nil {
		query += " LIMIT ?"
		args = append(args, *limit)
	}
	if offset != nil {
		query += " OFFSET ?"
		args = append(args, *offset)
	}
	rows, err = d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var game Game
		err := rows.Scan(&game.GameId, &game.UserId, &game.Name, &game.Publisher, &game.Year, &game.System, &game.Condition, &game.Owners)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return games, nil
}

func (d *SQLDatastore) CreateGame(game *Game) (*Game, error) {
	result, err := d.db.Exec("INSERT INTO games (`userId`, `name`, `publisher`, `year`, `system`, `condition`) VALUES (?, ?, ?, ?, ?, ?)", game.UserId, game.Name, game.Publisher, game.Year, game.System, *game.Condition)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	intId := int(id)
	game.GameId = &intId

	if game.Owners != nil {
		_, err = d.db.Exec("UPDATE games SET owners = ? WHERE `gameId` = ?", game.Owners, game.GameId)
		if err != nil {
			return nil, err
		}
	}

	return game, nil
}

func (d *SQLDatastore) UpdateGame(id int, game *Game) error {
	query := "UPDATE games SET "
	args := []interface{}{}
	updates := []string{}

	if game.Name != nil {
		updates = append(updates, "`name` = ?")
		args = append(args, game.Name)
	}

	if game.Publisher != nil {
		updates = append(updates, "`publisher` = ?")
		args = append(args, game.Publisher)
	}

	if game.Year != nil {
		updates = append(updates, "`year` = ?")
		args = append(args, game.Year)
	}

	if game.System != nil {
		updates = append(updates, "`system` = ?")
		args = append(args, game.System)
	}

	if game.Condition != nil {
		updates = append(updates, "`condition` = ?")
		args = append(args, game.Condition)
	}

	if game.Owners != nil {
		updates = append(updates, "`owners` = ?")
		args = append(args, game.Owners)
	}

	if len(updates) == 0 {
		return nil
	}

	query += strings.Join(updates, ", ")
	query += " WHERE `gameId` = ?"
	args = append(args, id)
	_, err := d.db.Exec(query, args...)
	return err
}

func (d *SQLDatastore) DeleteGame(id int) error {
	_, err := d.db.Exec("DELETE FROM games WHERE `gameId` = ?", id)
	return err
}

func (d *SQLDatastore) ChangeGameUserId(id int, userId int) error {
	_, err := d.db.Exec("UPDATE games SET `userId` = ? WHERE `gameId` = ?", userId, id)
	return err
}

// ------------------- Offers -------------------//
func (d *SQLDatastore) GetOffer(id int) (*Offer, error) {
	var offer Offer
	err := d.db.QueryRow("SELECT * FROM offers WHERE `offerId` = ?", id).Scan(&offer.OfferId, &offer.OffererUserId, &offer.RecipientUserId, &offer.OffererGameId, &offer.RecipientGameId, &offer.Status)
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (d *SQLDatastore) GetOffers(offererUserId *int, recipientUserId *int, limit *int, offset *int) ([]Offer, error) {
	var offers []Offer
	var rows *sql.Rows
	var err error
	query := "SELECT * FROM offers"
	args := []interface{}{}

	if offererUserId != nil {
		query += " WHERE `offererUserId` = ?"
		args = append(args, *offererUserId)
	}
	if recipientUserId != nil {
		query += " WHERE `recipientUserId` = ?"
		args = append(args, *recipientUserId)
	}
	if limit != nil {
		query += " LIMIT ?"
		args = append(args, *limit)
	}
	if offset != nil {
		query += " OFFSET ?"
		args = append(args, *offset)
	}
	rows, err = d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var offer Offer
		err := rows.Scan(&offer.OfferId, &offer.OffererUserId, &offer.RecipientUserId, &offer.OffererGameId, &offer.RecipientGameId, &offer.Status)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func (d *SQLDatastore) CreateOffer(offer *Offer) (*Offer, error) {
	result, err := d.db.Exec("INSERT INTO offers (`offererUserId`, `recipientUserId`, `offererGameId`, `recipientGameId`, `status`) VALUES (?, ?, ?, ?, ?)", offer.OffererUserId, offer.RecipientUserId, offer.OffererGameId, offer.RecipientGameId, offer.Status)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	intId := int(id)
	offer.OfferId = &intId

	return offer, nil
}

func (d *SQLDatastore) UpdateOffer(id int, offer *Offer) error {
	query := "UPDATE offers SET "
	args := []interface{}{}
	updates := []string{}

	updates = append(updates, "`status` = ?")
	args = append(args, offer.Status)

	if len(updates) == 0 {
		return nil
	}

	query += strings.Join(updates, ", ")
	query += " WHERE `offerId` = ?"
	args = append(args, id)
	_, err := d.db.Exec(query, args...)
	return err
}

func (d *SQLDatastore) DeleteOffer(id int) error {
	_, err := d.db.Exec("DELETE FROM offers WHERE `offerId` = ?", id)
	return err
}
