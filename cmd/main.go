package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	"github.com/LLIEPJIOK/calculating-server/internal/threadInfo"
	"github.com/gorilla/mux"
)

var (
	numberOfThreads = 10
	expressionsChan = make(chan expression.Expression, 1000)
	threadInfos     = make([]*threadInfo.ThreadInfo, numberOfThreads)

	configurationFuncMap = template.FuncMap{
		"GetOperationTime": expression.GetOperationTime,
	}

	inputExpressionTemplate    = template.Must(template.ParseFiles("web/templates/inputExpressionTemplate.html"))
	inputListTemplate          = template.Must(template.ParseFiles("web/templates/inputListTemplate.html"))
	listExpressionsTemplate    = template.Must(template.ParseFiles("web/templates/listExpressionsTemplate.html"))
	configurationTemplate      = template.Must(template.New("configurationTemplate.html").Funcs(configurationFuncMap).ParseFiles("web/templates/configurationTemplate.html"))
	computingResourcesTemplate = template.Must(template.ParseFiles("web/templates/computingResourcesTemplate.html"))
)

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
	database.InsertExpressionInBD(&exp)

	if err != nil {
		exp.Status = err.Error()
	} else {
		exp.Status = "in queue"
		expressionsChan <- exp
	}

	database.UpdateStatus(&exp)

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
	if err := computingResourcesTemplate.Execute(writer, threadInfos); err != nil {
		log.Printf("inputExpressionTemplate error: %v\n", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

func threadFunc(threadInfo *threadInfo.ThreadInfo) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			threadInfo.UpdateStatus("Waiting for expression")
		case exp, ok := <-expressionsChan:
			if !ok {
				threadInfo.UpdateStatus("Closed")
				return
			}

			threadInfo.UpdateStatus(fmt.Sprintf("Calculation expression #%v", exp.Id))
			exp.Status = "calculating"
			database.UpdateStatus(&exp)

			exp.Calculate()

			database.UpdateStatus(&exp)
			database.UpdateResult(&exp)
			threadInfo.UpdateStatus("Waiting for expression")
		}
	}
}

func init() {
	for i := range numberOfThreads {
		threadInfos[i] = &threadInfo.ThreadInfo{
			LastPing: time.Now(),
			Status:   "Waiting for expression",
			Id:       i + 1,
		}
		go threadFunc(threadInfos[i])
	}

	expressions := database.GetUncalculatingExpressions()
	for _, expression := range expressions {
		_ = expression.Parse()
		expressionsChan <- *expression
	}
}

func main() {
	defer database.Close()
	defer close(expressionsChan)

	router := mux.NewRouter()
	router.Use(RecoverMiddleware)
	router.HandleFunc("/", InputExpressionHandler).Methods("GET")
	router.HandleFunc("/add-expression", AddExpressionHandler).Methods("POST")
	router.HandleFunc("/list-expressions", ListExpressionsHandler).Methods("GET")
	router.HandleFunc("/configuration", ConfigurationHandler).Methods("GET")
	router.HandleFunc("/configuration/change", ChangeConfigurationHandler).Methods("PUT")
	router.HandleFunc("/computing-resources", ComputingResourcesHandler).Methods("GET")

	fs := http.FileServer(http.Dir("static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fs))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
