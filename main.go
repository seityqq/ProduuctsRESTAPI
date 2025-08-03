package main

import (
	"CRUDitems/api"
	"CRUDitems/internal/migrates"
	_ "database/sql"
	"encoding/json"
	"fmt"
	mlib "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var err error
	var dbInstance api.DataBase
	dbInstance = api.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer dbInstance.DB.Close()

	file, err := os.OpenFile("productsLog.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("File Open Error")
	}
	defer file.Close()
	log.SetOutput(file)

	migrates.RunMigrations("postgres://postgres:postgres@localhost/postgres?sslmode=disable")

	r := mux.NewRouter()

	// GETALLPRODUCTS GET
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		products, err := api.NewModificateDB().GetAllProducts(dbInstance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	}).Methods("GET")

	// CREATE POST
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		var product api.Product

		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		createdProduct, err := api.NewModificateDB().CreateProduct(product, dbInstance)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdProduct)
	}).Methods("POST")

	// UPDATE PUT
	r.HandleFunc("/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var product api.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		product.ID = id

		updatedProduct, err := api.NewModificateDB().UpdateProduct(product, dbInstance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(updatedProduct)
	}).Methods("PUT")

	// ROLLBACK POST
	r.HandleFunc("/products/{rollbackId}/rollback", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		rollbackID, err := strconv.Atoi(params["rollbackId"])
		if err != nil {
			http.Error(w, "Invalid rollback ID", http.StatusBadRequest)
			return
		}

		err = api.NewModificateDB().RollbackProduct(uint(rollbackID))
		if err != nil && err != mlib.ErrNoChange {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		version, _ := mlib.New("file://migrations", api.DSN)
		v, _, _ := version.Version()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Rolled back to version %d\n", v)))
	}).Methods("POST")

	// GETPRODUCTHISTORY GET
	r.HandleFunc("/products/{id}/history", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		history, err := api.NewModificateDB().GetProductHistory(id, dbInstance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}).Methods("GET")

	// ROLLBACKPRODUCTTOVERSION POST
	r.HandleFunc("/products/{id}/{historyId}/rollback", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		historyId, err := strconv.Atoi(params["historyId"])
		if err != nil {
			http.Error(w, "Invalid history ID", http.StatusBadRequest)
			return
		}

		rolledBackProduct, err := api.NewModificateDB().RollbackProductToVersion(id, historyId, dbInstance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rolledBackProduct)
	}).Methods("POST")

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
