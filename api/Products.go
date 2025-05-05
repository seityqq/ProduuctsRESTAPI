package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

const (
	dsn = "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
)

type Product struct {
	id          int     `json:id`
	label       string  `json:label`
	description string  `json:description`
	price       float64 `json:price`
}

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "postgres")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Database connection failed: %s", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database ping failed: %s", err)
	}

	fmt.Println("Successfully connected to database")
	return db, nil
}

func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	_, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rows, err := DB.Query("SELECT * FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.id, &product.label, &product.description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	json.NewEncoder(w).Encode(products)
}
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sqlStatement := `INSERT INTO products(label, description, price) VALUES ($1, $2, $3)`
	_, err = DB.Exec(sqlStatement, product.label, product.description, product.price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	_, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var product Product
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sqlStatement := `UPDATE products SET label = $1, description = &2, price = $3`
	_, err = DB.Exec(sqlStatement, product.label, product.description, product.price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RollbackProduct(w http.ResponseWriter, r *http.Request) {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		http.Error(w, "Error creating migrator: %v", http.StatusInternalServerError)
		return
	}

	if err := m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			w.Write([]byte("Nothing to rollback"))
			return
		}
		http.Error(w, fmt.Sprintf("Error creating migrator: %v", err), http.StatusInternalServerError)
		return
	}

	version, _, _ := m.Version()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Migration has been rolled back. New version: %d\n", version)))
}
