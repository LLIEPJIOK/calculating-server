package database

import (
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/user"
	_ "github.com/lib/pq"
)

func createUsersTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "users" (
			login VARCHAR(100) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			hash_password VARCHAR(100) NOT NULL
		)`)
	if err != nil {
		log.Fatal("error creating users table:", err)
	}
}

func InsertUser(insertingUser *user.User) {
	_, err := dataBase.Exec(`
		INSERT INTO "users"(login, name, hash_password) 
		VALUES($1, $2, $3)`,
		insertingUser.Login, insertingUser.Name, insertingUser.HashPassword)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *insertingUser, err)
		return
	}
}

func GetUserByLogin(login string) user.User {
	var gettingUser user.User
	dataBase.QueryRow(`
		SELECT * 
		FROM "users" 
		WHERE login = $1`, login).Scan(&gettingUser.Login, &gettingUser.Name, &gettingUser.HashPassword)

	return gettingUser
}
