package dal

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type SQLDatastore struct {
	db *sql.DB
}

func Init(host string, port string, user string, password string, protocol string, dbName string) (*SQLDatastore, error) {
	d := new(SQLDatastore)
	cfg := mysql.Config{
		User:   user,
		Passwd: password,
		Net:    protocol,
		Addr:   host + ":" + port,
		DBName: dbName,
	}
	fmt.Printf("Connecting to database with %s\n", cfg.FormatDSN())
	var err error
	d.db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	connected := false
	for attempts := 0; attempts < 5; attempts++ {
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
	return d, nil
}

func (d *SQLDatastore) Close() error {
	return d.db.Close()
}

func (d *SQLDatastore) GetOfferDetails(offerId int) (*Offer, error) {
	var offer Offer
	err := d.db.QueryRow("SELECT `offererUserId`, `recipientUserId` FROM `offers` WHERE `offerId` = ?", offerId).Scan(&offer.OffererUserId, &offer.RecipientUserId)
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (d *SQLDatastore) GetUserDetails(userId int) (*User, error) {
	var user User
	err := d.db.QueryRow("SELECT `email`, `name`, `password` FROM `users` WHERE `userId` = ?", userId).Scan(&user.Email, &user.Name, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
