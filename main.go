package main

import (
	"calculating-server/expressions"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ThreadInfo struct {
	LastPing time.Time
	Status   string
	Id       int
}

func (ti *ThreadInfo) UpdateStatus(newStatus string) {
	ti.Status = newStatus
	ti.LastPing = time.Now()
}

var (
	numberOfThreads = 10
	db              = expressions.NewDB()
	expressionsChan = make(chan *expressions.Expression, 1000)
	threadInfos     = make([]*ThreadInfo, numberOfThreads)

	inputExpressionTemplate    = template.Must(template.ParseFiles("static/templates/inputExpressionTemplate.html"))
	inputListTemplate          = template.Must(template.ParseFiles("static/templates/inputListTemplate.html"))
	listExpressionsTemplate    = template.Must(template.ParseFiles("static/templates/listExpressionsTemplate.html"))
	configurationTemplate      = template.Must(template.ParseFiles("static/templates/configurationTemplate.html"))
	computingResourcesTemplate = template.Must(template.ParseFiles("static/templates/computingResourcesTemplate.html"))
)

func InputExpressionHandler(w http.ResponseWriter, r *http.Request) {
	if err := inputExpressionTemplate.Execute(w, db.LastInputs); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	input := r.PostFormValue("expression")
	exp, err := expressions.NewExpression(input)
	db.InsertExpressionInBD(exp)

	if err != nil {
		exp.Status = err.Error()
	} else {
		exp.Status = "in queue"
		expressionsChan <- exp
	}

	db.UpdateStatus(exp)

	if err := inputListTemplate.Execute(w, db.LastInputs); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ListExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	searchId := r.URL.Query().Get("id")
	exps := db.GetExpressionById(searchId)
	if err := listExpressionsTemplate.Execute(w, struct {
		Exps     []*expressions.Expression
		SearchId string
	}{
		Exps:     exps,
		SearchId: searchId,
	}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	if err := configurationTemplate.Execute(w, db); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ChangeConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	timePlus, _ := strconv.ParseInt(r.PostFormValue("time-plus"), 10, 64)
	timeMinus, _ := strconv.ParseInt(r.PostFormValue("time-minus"), 10, 64)
	timeMultiply, _ := strconv.ParseInt(r.PostFormValue("time-multiply"), 10, 64)
	timeDivide, _ := strconv.ParseInt(r.PostFormValue("time-divide"), 10, 64)

	db.UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide)
}

func ComputingResourcesHandler(w http.ResponseWriter, r *http.Request) {
	if err := computingResourcesTemplate.Execute(w, threadInfos); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

func threadFunc(threadInfo *ThreadInfo) {
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
			db.UpdateStatus(exp)

			exp.Calculate()

			db.UpdateStatus(exp)
			db.UpdateResult(exp)
			threadInfo.UpdateStatus("Waiting for expression")
		}
	}
}

func init() {
	for i := range numberOfThreads {
		threadInfos[i] = &ThreadInfo{
			LastPing: time.Now(),
			Status:   "Waiting for expression",
			Id:       i + 1,
		}
		go threadFunc(threadInfos[i])
	}

	exps := db.GetUncalculatingExpressions()
	for _, v := range exps {
		_ = v.Parse()
		expressionsChan <- v
	}
}

func main() {
	defer db.Close()
	defer close(expressionsChan)

	r := mux.NewRouter()
	r.Use(RecoverMiddleware)
	r.HandleFunc("/", InputExpressionHandler).Methods("GET")
	r.HandleFunc("/add-expression", AddExpressionHandler).Methods("POST")
	r.HandleFunc("/list-expressions", ListExpressionsHandler).Methods("GET")
	r.HandleFunc("/configuration", ConfigurationHandler).Methods("GET")
	r.HandleFunc("/configuration/change", ChangeConfigurationHandler).Methods("PUT")
	r.HandleFunc("/computing-resources", ComputingResourcesHandler).Methods("GET")

	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fs))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
