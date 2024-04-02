package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	"github.com/LLIEPJIOK/calculating-server/internal/user"
	"github.com/LLIEPJIOK/calculating-server/internal/workers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type userContextKey string

const (
	secretString                 = "super_secret_string"
	keyUserString userContextKey = "user"
)

var (
	configurationFuncMap = template.FuncMap{
		"GetOperationTime": expression.GetOperationTime,
	}

	loginTemplate              = template.Must(template.ParseFiles("web/templates/login.html"))
	registerTemplate           = template.Must(template.ParseFiles("web/templates/register.html"))
	invalidLoginRegister       = template.Must(template.ParseFiles("web/templates/invalidLoginRegister.html"))
	inputExpressionTemplate    = template.Must(template.ParseFiles("web/templates/inputExpression.html"))
	inputListTemplate          = template.Must(template.ParseFiles("web/templates/inputList.html"))
	listExpressionsTemplate    = template.Must(template.ParseFiles("web/templates/listExpressions.html"))
	configurationTemplate      = template.Must(template.New("configuration.html").Funcs(configurationFuncMap).ParseFiles("web/templates/configuration.html"))
	computingResourcesTemplate = template.Must(template.ParseFiles("web/templates/computingResources.html"))
)

func CheckingTokenMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie("Authorization")
		if errors.Is(http.ErrNoCookie, err) {
			fmt.Println("op")
			next.ServeHTTP(writer, request)
			return
		}
		if err != nil {
			log.Printf("getting cookie error: %v\n", err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		tokenString := cookie.Value
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(secretString), nil
		})
		if err != nil {
			log.Printf("error parsing token: %v\n", err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !token.Valid {
			log.Printf("token is invalid: %#v\n", token)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("cannot cast token claim to MapClaims")
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			log.Println("token is expired")
			next.ServeHTTP(writer, request)
			return
		}
		ctx := context.WithValue(request.Context(), keyUserString, database.GetUserByLogin(claims["login"].(string)))
		http.Redirect(writer, request.WithContext(ctx), "http://localhost:8081/input-expression", http.StatusSeeOther)
	})

}

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

func LoginHandler(writer http.ResponseWriter, request *http.Request) {
	if err := loginTemplate.Execute(writer, nil); err != nil {
		log.Printf("loginTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("op")
	if err := registerTemplate.Execute(writer, nil); err != nil {
		log.Printf("registerTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ConfirmRegistrationHandler(writer http.ResponseWriter, request *http.Request) {
	name := request.PostFormValue("name")
	login := request.PostFormValue("login")
	password := request.PostFormValue("password")

	checkingUser := database.GetUserByLogin(login)
	if checkingUser.Login != "" {
		invalidLoginRegister.Execute(writer, nil)
		return
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Printf("error while hashing password: %v\n", err)
	}

	database.InsertUserInDatabase(&user.User{Login: login, Name: name, HashPassword: string(hashPassword)})
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"nbf":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"iat":   time.Now(),
	})

	tokenString, err := token.SignedString([]byte(secretString))
	if err != nil {
		log.Printf("generate token error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})
}

func InputExpressionHandler(writer http.ResponseWriter, request *http.Request) {
	if err := inputExpressionTemplate.Execute(writer, database.GetLastExpressions()); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func AddExpressionHandler(writer http.ResponseWriter, request *http.Request) {
	input := request.PostFormValue("expression")
	exp, err := expression.NewExpression(input)
	database.InsertExpressionInDatabase(&exp)

	if err != nil {
		exp.Status = err.Error()
	} else {
		exp.Status = "in queue"
		workers.ExpressionsChan <- exp
	}

	database.UpdateExpressionStatus(&exp)

	if err := inputListTemplate.Execute(writer, database.GetLastExpressions()); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ListExpressionsHandler(writer http.ResponseWriter, request *http.Request) {
	searchId := request.URL.Query().Get("id")
	exps := database.GetExpressionById(searchId)
	if err := listExpressionsTemplate.Execute(writer, struct {
		Exps     []*expression.Expression
		SearchId string
	}{
		Exps:     exps,
		SearchId: searchId,
	}); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ConfigurationHandler(writer http.ResponseWriter, request *http.Request) {
	if err := configurationTemplate.Execute(writer, expression.GetOperationTimes()); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ChangeConfigurationHandler(writer http.ResponseWriter, request *http.Request) {
	timePlus, _ := strconv.ParseInt(request.PostFormValue("time-plus"), 10, 64)
	timeMinus, _ := strconv.ParseInt(request.PostFormValue("time-minus"), 10, 64)
	timeMultiply, _ := strconv.ParseInt(request.PostFormValue("time-multiply"), 10, 64)
	timeDivide, _ := strconv.ParseInt(request.PostFormValue("time-divide"), 10, 64)

	database.UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide)
}

func ComputingResourcesHandler(writer http.ResponseWriter, request *http.Request) {
	if err := computingResourcesTemplate.Execute(writer, workers.Workers); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ConfigureControllers(router *mux.Router) {
	router.Use(RecoverMiddleware)
	router.HandleFunc("/", LoginHandler).Methods("GET")
	router.HandleFunc("/register", CheckingTokenMiddleWare(RegisterHandler)).Methods("GET")
	router.HandleFunc("/register/confirm", ConfirmRegistrationHandler).Methods("POST")
	router.HandleFunc("/input-expression", InputExpressionHandler).Methods("GET")
	router.HandleFunc("/add-expression", AddExpressionHandler).Methods("POST")
	router.HandleFunc("/list-expressions", ListExpressionsHandler).Methods("GET")
	router.HandleFunc("/configuration", ConfigurationHandler).Methods("GET")
	router.HandleFunc("/configuration/change", ChangeConfigurationHandler).Methods("PUT")
	router.HandleFunc("/computing-resources", ComputingResourcesHandler).Methods("GET")

	fileServer := http.FileServer(http.Dir("./web/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileServer))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8081", router))
}
