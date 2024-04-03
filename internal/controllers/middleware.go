package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

const (
	keyUserString userContextKey = "user"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(writer, fmt.Sprintf("panic: %v", err), http.StatusInternalServerError)
				log.Println("recovering from panic:", err)
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

func getUserFromTokenString(tokenString string) (user.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretString), nil
	})
	if err != nil {
		return user.User{}, fmt.Errorf("error parsing token: %v", err)
	}
	if !token.Valid {
		return user.User{}, fmt.Errorf("token is invalid: %#v", token)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return user.User{}, fmt.Errorf("cannot cast token claim to MapClaims")
	}
	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		return user.User{}, fmt.Errorf("token is expired")
	}
	return database.GetUserByLogin(claims["login"].(string)), nil
}

func CheckingTokenBeforeLoginMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie("Authorization")
		if errors.Is(http.ErrNoCookie, err) {
			next.ServeHTTP(writer, request)
			return
		}
		if err != nil {
			log.Printf("getting cookie error: %v\n", err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		tokenString := cookie.Value
		currentUser, err := getUserFromTokenString(tokenString)
		if err != nil {
			log.Println(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(request.Context(), keyUserString, currentUser)
		http.Redirect(writer, request.WithContext(ctx), "/input-expression", http.StatusSeeOther)
	})
}

func CheckingTokenAfterLoginMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie("Authorization")
		if errors.Is(http.ErrNoCookie, err) {
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}
		if err != nil {
			log.Printf("getting cookie error: %v\n", err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		tokenString := cookie.Value
		currentUser, err := getUserFromTokenString(tokenString)
		if err != nil {
			log.Println(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(request.Context(), keyUserString, currentUser)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
