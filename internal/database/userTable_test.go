package database

import (
	"os"
	"strconv"
	"testing"

	"github.com/LLIEPJIOK/calculating-server/internal/user"
)

func TestUserTable(t *testing.T) {
	users := []user.User{
		{
			Login:        "1",
			Name:         "Denis",
			HashPassword: "2345",
		},
		{
			Login:        "2",
			Name:         "Matvey",
			HashPassword: "n5kj34jlb",
		},
		{
			Login:        "3",
			Name:         "Vova",
			HashPassword: "jkl4kln",
		},
		{
			Login:        "4",
			Name:         "Dasha",
			HashPassword: "mrejefw;j",
		},
		{
			Login:        "5",
			Name:         "Anton",
			HashPassword: "konb5k4nsdf3",
		},
		{
			Login:        "6",
			Name:         "Lera",
			HashPassword: "7483ienfr4",
		},
	}

	os.Setenv("expressionDatabaseName", "user_table_test_db")
	Configure()
	defer deleteDatabase()

	for _, user := range users {
		InsertUser(&user)
	}

	for i := 0; i < 6; i++ {
		user := GetUserByLogin(strconv.Itoa(i + 1))
		if user != users[i] {
			t.Fatalf("incorrect user: expected: %v, but got: %v", users[i], user)
		}
	}
}
