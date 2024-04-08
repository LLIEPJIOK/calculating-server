package controllers

import (
	"log"
	"net/http"
	"reflect"
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

type InputExpressionPageInfo struct {
	UserName    string
	Expressions []*expression.Expression
}

const (
	secretString                 = "super_secret_string"
	keyUserString userContextKey = "user"
)

var (
	configurationFuncMap = template.FuncMap{
		"GetOperationTime": expression.GetOperationTime,
	}

	logInTemplate              = template.Must(template.ParseFiles("static/templates/logIn.html"))
	logInFeedbackTemplate      = template.Must(template.ParseFiles("static/templates/logInFeedback.html"))
	registerTemplate           = template.Must(template.ParseFiles("static/templates/register.html"))
	registerFeedbackTemplate   = template.Must(template.ParseFiles("static/templates/registerFeedback.html"))
	inputExpressionTemplate    = template.Must(template.ParseFiles("static/templates/inputExpression.html"))
	inputListTemplate          = template.Must(template.ParseFiles("static/templates/inputList.html"))
	listExpressionsTemplate    = template.Must(template.ParseFiles("static/templates/listExpressions.html"))
	configurationTemplate      = template.Must(template.New("configuration.html").Funcs(configurationFuncMap).ParseFiles("static/templates/configuration.html"))
	computingResourcesTemplate = template.Must(template.ParseFiles("static/templates/computingResources.html"))
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
	if err := logInTemplate.Execute(writer, nil); err != nil {
		log.Printf("loginTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func returnLogInFeedback(writer http.ResponseWriter, registerFeedback *RegisterFeedback) {
	if err := logInFeedbackTemplate.Execute(writer, registerFeedback); err != nil {
		log.Printf("logInFeedbackTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func ConfirmLogInHandler(writer http.ResponseWriter, request *http.Request) {
	login := request.PostFormValue("login")
	password := request.PostFormValue("password")

	registerFeedback := &RegisterFeedback{
		LoginValue:    login,
		PasswordValue: password,
	}

	checkingUser := database.GetUserByLogin(login)
	if err := bcrypt.CompareHashAndPassword([]byte(checkingUser.HashPassword), []byte(password)); err != nil || password == "" {
		registerFeedback.LoginFeedback = "Invalid login or password"
		registerFeedback.PasswordFeedback = "Invalid login or password"
		writer.WriteHeader(http.StatusNonAuthoritativeInfo)
		returnLogInFeedback(writer, registerFeedback)
		return
	}

	generateAndReturnToken(writer, login)
	returnLogInFeedback(writer, registerFeedback)
}

func RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	if err := registerTemplate.Execute(writer, nil); err != nil {
		log.Printf("registerTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func returnRegisterFeedback(writer http.ResponseWriter, registerFeedback *RegisterFeedback) {
	if err := registerFeedbackTemplate.Execute(writer, registerFeedback); err != nil {
		log.Printf("registerFeedbackTemplate error: %v\n", err)
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
		writer.WriteHeader(http.StatusNonAuthoritativeInfo)
		returnRegisterFeedback(writer, registerFeedback)
		return
	}

	checkingUser := database.GetUserByLogin(login)
	if checkingUser.Login != "" {
		registerFeedback.LoginFeedback = "User with such login already exists"
		writer.WriteHeader(http.StatusNonAuthoritativeInfo)
		returnRegisterFeedback(writer, registerFeedback)
		return
	}

	if registerFeedback.NameFeedback != "" || registerFeedback.PasswordFeedback != "" || password != repeatPassword {
		writer.WriteHeader(http.StatusNonAuthoritativeInfo)
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
	database.InsertDefaultOperationTimes(login)
	generateAndReturnToken(writer, login)
	returnRegisterFeedback(writer, registerFeedback)
}

func LogOutHandler(writer http.ResponseWriter, request *http.Request) {
	http.SetCookie(writer, &http.Cookie{
		Name:   "Authorization",
		MaxAge: -1,
	})
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func InputExpressionHandler(writer http.ResponseWriter, request *http.Request) {
	contextUser := request.Context().Value(keyUserString)
	currentUser, ok := contextUser.(user.User)
	if !ok {
		log.Printf("InputExpressionHandler: expected: user, but found: %v\n", reflect.TypeOf(contextUser))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := inputExpressionTemplate.Execute(writer, InputExpressionPageInfo{
		UserName:    currentUser.Name,
		Expressions: database.GetLastExpressions(currentUser.Login),
	}); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func AddExpressionHandler(writer http.ResponseWriter, request *http.Request) {
	contextUser := request.Context().Value(keyUserString)
	currentUser, ok := contextUser.(user.User)
	if !ok {
		log.Printf("AddExpressionHandler: expected: user, but found: %v\n", reflect.TypeOf(contextUser))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	input := request.PostFormValue("expression")
	exp := expression.New(currentUser.Login, input)
	operationsTime, err := database.GetOperationsTime(currentUser.Login)
	if err == nil {
		exp.OperationsTimes = operationsTime
	}
	if exp.Status == "" {
		exp.Status = "in queue"
	}
	database.InsertExpression(&exp)
	if exp.Status == "in queue" {
		workers.ExpressionsChan <- exp
	}

	if err := inputListTemplate.Execute(writer, database.GetLastExpressions(currentUser.Login)); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ListExpressionsHandler(writer http.ResponseWriter, request *http.Request) {
	contextUser := request.Context().Value(keyUserString)
	currentUser, ok := contextUser.(user.User)
	if !ok {
		log.Printf("ListExpressionsHandler: expected: user, but found: %v\n", reflect.TypeOf(contextUser))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	searchId := request.URL.Query().Get("id")
	exps := database.GetExpressionsById(searchId, currentUser.Login)
	// TODO: create struct
	if err := listExpressionsTemplate.Execute(writer, struct {
		UserName string
		Exps     []*expression.Expression
		SearchId string
	}{
		UserName: currentUser.Name,
		Exps:     exps,
		SearchId: searchId,
	}); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ConfigurationHandler(writer http.ResponseWriter, request *http.Request) {
	contextUser := request.Context().Value(keyUserString)
	currentUser, ok := contextUser.(user.User)
	if !ok {
		log.Printf("ConfigurationHandler: expected: user, but found: %v\n", reflect.TypeOf(contextUser))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	operationsTimes, err := database.GetOperationsTime(currentUser.Login)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if err := configurationTemplate.Execute(writer, struct {
		UserName       string
		OperationsTime map[string]uint64
	}{
		UserName:       currentUser.Name,
		OperationsTime: operationsTimes,
	}); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ChangeConfigurationHandler(writer http.ResponseWriter, request *http.Request) {
	contextUser := request.Context().Value(keyUserString)
	currentUser, ok := contextUser.(user.User)
	if !ok {
		log.Printf("ChangeConfigurationHandler: expected: user, but found: %v\n", reflect.TypeOf(contextUser))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	timePlus, _ := strconv.ParseUint(request.PostFormValue("time-plus"), 10, 64)
	timeMinus, _ := strconv.ParseUint(request.PostFormValue("time-minus"), 10, 64)
	timeMultiply, _ := strconv.ParseUint(request.PostFormValue("time-multiply"), 10, 64)
	timeDivide, _ := strconv.ParseUint(request.PostFormValue("time-divide"), 10, 64)

	database.UpdateOperationTimes(timePlus, timeMinus, timeMultiply, timeDivide, currentUser.Login)
}

func ComputingResourcesHandler(writer http.ResponseWriter, request *http.Request) {
	contextUser := request.Context().Value(keyUserString)
	currentUser, ok := contextUser.(user.User)
	if !ok {
		log.Printf("ComputingResourcesHandler: expected: user, but found: %v\n", reflect.TypeOf(contextUser))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := computingResourcesTemplate.Execute(writer, struct {
		UserName string
		Workers  []*workers.Worker
	}{
		UserName: currentUser.Name,
		Workers:  workers.Workers,
	}); err != nil {
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
	router.HandleFunc("/log-out", LogOutHandler).Methods("GET")
	router.HandleFunc("/input-expression", CheckingTokenAfterLoginMiddleWare(InputExpressionHandler)).Methods("GET")
	router.HandleFunc("/add-expression", CheckingTokenAfterLoginMiddleWare(AddExpressionHandler)).Methods("POST")
	router.HandleFunc("/list-expressions", CheckingTokenAfterLoginMiddleWare(ListExpressionsHandler)).Methods("GET")
	router.HandleFunc("/configuration", CheckingTokenAfterLoginMiddleWare(ConfigurationHandler)).Methods("GET")
	router.HandleFunc("/configuration/change", CheckingTokenAfterLoginMiddleWare(ChangeConfigurationHandler)).Methods("PUT")
	router.HandleFunc("/computing-resources", CheckingTokenAfterLoginMiddleWare(ComputingResourcesHandler)).Methods("GET")

	fileServer := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileServer))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
