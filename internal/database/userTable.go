package database

import (
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/user"
	_ "github.com/lib/pq"
)

func InsertUser(insertingUser *user.User) {
	_, err := dataBase.Exec(`
		INSERT INTO "User"(login, name, hash_password) 
		VALUES($1, $2, $3)
		`, insertingUser.Login, insertingUser.Name, insertingUser.HashPassword)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *insertingUser, err)
		return
	}
}

func GetUserByLogin(login string) user.User {
	var gettingUser user.User
	dataBase.QueryRow(`
		SELECT
			login, name, hash_password 
		FROM "User"
		WHERE login = $1`, login).Scan(&gettingUser.Login, &gettingUser.Name, &gettingUser.HashPassword)

	return gettingUser
}

func createUsersTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "User" (
			login TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			hash_password TEXT NOT NULL
		)`)
	if err != nil {
		log.Fatal("error creating users table:", err)
	}
}
