package controllers

import (
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

type RegisterFeedback struct {
	NameValue              string
	NameFeedback           string
	LoginValue             string
	LoginFeedback          string
	PasswordValue          string
	PasswordFeedback       string
	RepeatPasswordValue    string
	RepeatPasswordFeedback string
}

const (
	secretString = "super_secret_string"
)

var (
	configurationFuncMap = template.FuncMap{
		"GetOperationTime": expression.GetOperationTime,
	}

	loginTemplate               = template.Must(template.ParseFiles("static/templates/logIn.html"))
	registerTemplate            = template.Must(template.ParseFiles("static/templates/register.html"))
	invalidRegisterDataTemplate = template.Must(template.ParseFiles("static/templates/invalidRegisterData.html"))
	inputExpressionTemplate     = template.Must(template.ParseFiles("static/templates/inputExpression.html"))
	inputListTemplate           = template.Must(template.ParseFiles("static/templates/inputList.html"))
	listExpressionsTemplate     = template.Must(template.ParseFiles("static/templates/listExpressions.html"))
	configurationTemplate       = template.Must(template.New("configuration.html").Funcs(configurationFuncMap).ParseFiles("static/templates/configuration.html"))
	computingResourcesTemplate  = template.Must(template.ParseFiles("static/templates/computingResources.html"))
)

func generateAndReturnToken(writer http.ResponseWriter, login string) {
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

func LogInHandler(writer http.ResponseWriter, request *http.Request) {
	if err := loginTemplate.Execute(writer, nil); err != nil {
		log.Printf("loginTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ConfirmLogInHandler(writer http.ResponseWriter, request *http.Request) {
	login := request.PostFormValue("login")
	password := request.PostFormValue("password")

	checkingUser := database.GetUserByLogin(login)
	if err := bcrypt.CompareHashAndPassword([]byte(checkingUser.HashPassword), []byte(password)); err != nil {
		// TODO: write incorrect login or password response
		return
	}

	generateAndReturnToken(writer, login)
}

func RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	if err := registerTemplate.Execute(writer, nil); err != nil {
		log.Printf("registerTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func returnRegisterFeedback(writer http.ResponseWriter, registerFeedback *RegisterFeedback) {
	writer.WriteHeader(http.StatusNonAuthoritativeInfo)
	if err := invalidRegisterDataTemplate.Execute(writer, registerFeedback); err != nil {
		log.Printf("invalidRegisterDataTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func ConfirmRegistrationHandler(writer http.ResponseWriter, request *http.Request) {
	name := request.PostFormValue("name")
	login := request.PostFormValue("login")
	password := request.PostFormValue("password")
	repeatPassword := request.PostFormValue("repeat_password")

	registerFeedback := &RegisterFeedback{
		NameValue:           name,
		LoginValue:          login,
		PasswordValue:       password,
		RepeatPasswordValue: repeatPassword,
	}

	if name == "" {
		registerFeedback.NameFeedback = "Please provide a valid name"
	}
	if password == "" {
		registerFeedback.PasswordFeedback = "Please provide a valid password"
	}
	if repeatPassword == "" {
		registerFeedback.RepeatPasswordFeedback = "Please provide a valid password"
	}
	if password != "" && repeatPassword != "" && password != repeatPassword {
		registerFeedback.PasswordFeedback = "Password mismatch"
		registerFeedback.RepeatPasswordFeedback = "Password mismatch"
	}
	if login == "" {
		registerFeedback.LoginFeedback = "Please provide a valid login"
		returnRegisterFeedback(writer, registerFeedback)
		return
	}

	checkingUser := database.GetUserByLogin(login)
	if checkingUser.Login != "" {
		registerFeedback.LoginFeedback = "User with such login already exists"
		returnRegisterFeedback(writer, registerFeedback)
		return
	}

	if registerFeedback.NameFeedback != "" || registerFeedback.PasswordFeedback != "" || password != repeatPassword {
		returnRegisterFeedback(writer, registerFeedback)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Printf("error while hashing password: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	database.InsertUser(&user.User{Login: login, Name: name, HashPassword: string(hashPassword)})
	generateAndReturnToken(writer, login)
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
	database.InsertExpression(&exp)

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

	database.UpdateOperationTimes(timePlus, timeMinus, timeMultiply, timeDivide)
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
	router.HandleFunc("/", CheckingTokenBeforeLoginMiddleWare(LogInHandler)).Methods("GET")
	router.HandleFunc("/login/confirm", ConfirmLogInHandler).Methods("POST")
	router.HandleFunc("/register", CheckingTokenBeforeLoginMiddleWare(RegisterHandler)).Methods("GET")
	router.HandleFunc("/register/confirm", ConfirmRegistrationHandler).Methods("POST")
	router.HandleFunc("/input-expression", CheckingTokenAfterLoginMiddleWare(InputExpressionHandler)).Methods("GET")
	router.HandleFunc("/add-expression", AddExpressionHandler).Methods("POST")
	router.HandleFunc("/list-expressions", CheckingTokenAfterLoginMiddleWare(ListExpressionsHandler)).Methods("GET")
	router.HandleFunc("/configuration", CheckingTokenAfterLoginMiddleWare(ConfigurationHandler)).Methods("GET")
	router.HandleFunc("/configuration/change", ChangeConfigurationHandler).Methods("PUT")
	router.HandleFunc("/computing-resources", CheckingTokenAfterLoginMiddleWare(ComputingResourcesHandler)).Methods("GET")

	fileServer := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileServer))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
