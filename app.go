package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func (app *App) Init() {
	app.Router = mux.NewRouter()

	// Routes
	app.Router.HandleFunc("/", test).Methods("GET")
}

func (app *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func test(w http.ResponseWriter, r *http.Request) {
	http.Get("http://127.0.0.1:80")
	log.Println("Test is going great")
}
