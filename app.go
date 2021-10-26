package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func (app *App) Init() {
	app.Router = mux.NewRouter()

	// Routes
	app.Router.HandleFunc("/distribution", test).Methods("POST")
	app.Router.HandleFunc("/generate/{number}", addOrders).Methods("POST")
}

func (app *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func test(w http.ResponseWriter, r *http.Request) {
	var order Order

	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	FinishedOrders <- order
}

func addOrders(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	number, _ := strconv.Atoi(params["number"])
	atomic.AddInt64(&OrdersNumber, int64(number))

}
