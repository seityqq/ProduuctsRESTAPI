package main

import (
	"CRUDitems/api"
	"CRUDitems/internal/migrate"

	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	var err error
	api.DB, err = api.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer api.DB.Close()

	file, err := os.OpenFile("productsLog.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("File Open Error")
	}
	defer file.Close()
	log.SetOutput(file)

	migrate.RunMigrations("postgres://postgres:postgres@localhost/postgres?sslmode=disable")

	r := mux.NewRouter()

	r.HandleFunc("/products", api.GetAllProducts).Methods("GET")
	r.HandleFunc("/products", api.CreateProduct).Methods("POST")
	r.HandleFunc("/products/{id}", api.UpdateProduct).Methods("PUT")
	r.HandleFunc("/products/{id}/rollback", api.RollbackProduct).Methods("POST")

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
