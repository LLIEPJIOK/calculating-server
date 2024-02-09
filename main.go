package main

import (
	"calculating-server/expressions"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ExpressionKey string

const (
	keyExpression ExpressionKey = "expression"
)

var (
	db              = expressions.NewDB()
	expressionsChan = make(chan *expressions.Expression)

	inputExpressionTemplate = template.Must(template.ParseFiles("static/templates/inputExpressionTemplate.html"))
	inputListTemplate       = template.Must(template.ParseFiles("static/templates/inputListTemplate.html"))
	listExpressionsTemplate = template.Must(template.ParseFiles("static/templates/listExpressionsTemplate.html"))
	configurationTemplate   = template.Must(template.ParseFiles("static/templates/configurationTemplate.html"))
)

func InputExpressionHandler(w http.ResponseWriter, r *http.Request) {
	inputExpressionTemplate.Execute(w, db.LastInputs)
}

func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	exp, ok := r.Context().Value(keyExpression).(*expressions.Expression)
	if !ok {
		http.Error(w, "expression not found in context", http.StatusInternalServerError)
		return
	}
	db.InsertExpressionInBD(exp)
	expressionsChan <- exp
	exp.Status = "in queue"
}

func ListExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressions := db.GetExpressionById(r.URL.Query().Get("id"))
	listExpressionsTemplate.Execute(w, expressions)
}

func ConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	configurationTemplate.Execute(w, db)
}

func ChangeConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	timePlus, _ := strconv.ParseInt(r.PostFormValue("time-plus"), 10, 64)
	timeMinus, _ := strconv.ParseInt(r.PostFormValue("time-minus"), 10, 64)
	timeMultiply, _ := strconv.ParseInt(r.PostFormValue("time-multiply"), 10, 64)
	timeDivide, _ := strconv.ParseInt(r.PostFormValue("time-divide"), 10, 64)

	db.UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide)
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, fmt.Sprintf("panic: %v", err), http.StatusInternalServerError)
				log.Println("recovering from panic:", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func ParsingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		input := r.PostFormValue("expression")
		exp, err := expressions.NewExpression(input)
		if err != nil {
			exp.Status = err.Error()
		} else {
			ctx := context.WithValue(r.Context(), keyExpression, exp)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		db.UpdateStatus(exp)
		inputListTemplate.Execute(w, db.LastInputs)
	})
}

func init() {
	for range 10 {
		go func() {
			for exp := range expressionsChan {
				exp.Status = "calculating"
				db.UpdateStatus(exp)

				exp.Calculate()

				db.UpdateStatus(exp)
				db.UpdateResult(exp)
			}
		}()
	}
}

func main() {
	defer db.Close()

	r := mux.NewRouter()
	r.Use(RecoverMiddleware)
	r.HandleFunc("/", InputExpressionHandler).Methods("GET")
	r.HandleFunc("/add-expression", ParsingMiddleware(AddExpressionHandler)).Methods("POST")
	r.HandleFunc("/list-expressions", ListExpressionsHandler).Methods("GET")
	r.HandleFunc("/configuration", ConfigurationHandler).Methods("GET")
	r.HandleFunc("/configuration/change", ChangeConfigurationHandler).Methods("PUT")

	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fs))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
