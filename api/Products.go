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
	"time"
)

const (
	dsn = "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
)

type Product struct {
	ID          int       `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "postgres")

	DB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Database connection failed: %s", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database ping failed: %s", err)
	}

	log.Println("Successfully connected to database")
	return DB, nil
}

func GetAllProducts(w http.ResponseWriter, r *http.Request) {

	rows, err := DB.Query("SELECT * FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Label, &product.Description, &product.Price, &product.CreatedAt)
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
	_, err = DB.Exec(sqlStatement, product.Label, product.Description, product.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	prodID, err := strconv.Atoi(params["id"])
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

	sqlStatement := `UPDATE products SET label = $1, description = $2, price = $3 WHERE id = $4`
	_, err = DB.Exec(sqlStatement, product.Label, product.Description, product.Price, prodID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RollbackProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	_, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rollbackID, err := strconv.Atoi(params["rollbackId"])
	if err != nil {
		http.Error(w, "Invalid rollback ID", http.StatusBadRequest)
		return
	}

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		http.Error(w, "Error creating migrator: %v", http.StatusInternalServerError)
		return
	}

	if err := m.Migrate(uint(rollbackID)); err != nil {
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
