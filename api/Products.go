package api

import (
	"database/sql"
	"fmt"
	mlib "github.com/golang-migrate/migrate/v4"
	"log"
	"net/http"
	"time"
)

const (
	DSN = "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
)

type ModificateDB interface {
	GetAllProducts(db DataBase) ([]Product, error)
	CreateProduct(product Product, DB DataBase) (Product, error)
	UpdateProduct(product Product, DB DataBase) (Product, error)
	RollbackProduct(toVersion uint) error
	RollbackProductToVersion(productID, historyID int, db DataBase) (Product, error)
	GetProductHistory(productID int, db DataBase) ([]Product, error)
}
type Product struct {
	ID          int       `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	CreatedAt   time.Time `json:"createtime"`
}

type ProductHistory struct {
	HistoryID   int       `json:"history_id"`
	ProductID   int       `json:"product_id"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	CreatedAt   time.Time `json:"createtime"`
	UpdatedAt   time.Time `json:"updatetime"`
}

type DataBase struct {
	DB *sql.DB
}

type Server struct {
	w http.ResponseWriter
	r *http.Request
}

type ProductServiceImpl struct{}

func NewModificateDB() ModificateDB {
	return &ProductServiceImpl{}
}

func InitDB() DataBase {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "postgres")

	conn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Database connection failed: %s", err)
		return DataBase{}
	}

	if err = conn.Ping(); err != nil {
		conn.Close()
		log.Fatal("Database ping failed: %s", err)
		return DataBase{}
	}

	log.Println("Successfully connected to database")
	return DataBase{DB: conn}
}

func (p *ProductServiceImpl) GetAllProducts(db DataBase) ([]Product, error) {

	rows, err := db.DB.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Label, &product.Description, &product.Price, &product.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (p *ProductServiceImpl) CreateProduct(product Product, db DataBase) (Product, error) {
	var id int
	var createdAt time.Time

	err := db.DB.QueryRow(
		"INSERT INTO products(label, description, price) VALUES ($1, $2, $3) RETURNING id, createtime",
		product.Label, product.Description, product.Price,
	).Scan(&id, &createdAt)

	if err != nil {
		return Product{}, err
	}

	product.ID = id
	product.CreatedAt = createdAt

	return product, nil
}

func (p *ProductServiceImpl) UpdateProduct(product Product, db DataBase) (Product, error) {

	tx, err := db.DB.Begin()
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
        INSERT INTO product_history (product_id, label, description, price)
        SELECT id, label, description, price FROM products WHERE id = $1
    `, product.ID)
	if err != nil {
		return Product{}, err
	}

	_, err = tx.Exec(`
        UPDATE products 
        SET label = $1, description = $2, price = $3 
        WHERE id = $4
    `, product.Label, product.Description, product.Price, product.ID)
	if err != nil {
		return Product{}, err
	}

	if err = tx.Commit(); err != nil {
		return Product{}, err
	}

	return product, nil
}

func (p *ProductServiceImpl) RollbackProduct(toVersion uint) error {
	m, err := mlib.New("file://migrations", DSN)
	if err != nil {
		return err
	}
	return m.Migrate(toVersion)
}

func (p *ProductServiceImpl) RollbackProductToVersion(productID, historyID int, db DataBase) (Product, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback()

	var product Product
	err = tx.QueryRow(`
        SELECT id, label, description, price, updated_at 
        FROM product_history 
        WHERE product_id = $1 AND id = $2
    `, productID, historyID).Scan(&product.ID, &product.Label, &product.Description, &product.Price, &product.CreatedAt)

	if err != nil {
		return Product{}, err
	}

	_, err = tx.Exec(`
        UPDATE products 
        SET label = $1, description = $2, price = $3 
        WHERE id = $4
    `, product.Label, product.Description, product.Price, productID)

	if err != nil {
		return Product{}, err
	}

	if err = tx.Commit(); err != nil {
		return Product{}, err
	}

	product.ID = productID
	return product, nil
}

func (p *ProductServiceImpl) GetProductHistory(productID int, db DataBase) ([]Product, error) {
	rows, err := db.DB.Query(`
        SELECT id, product_id, label, description, price, updated_at 
        FROM product_history 
        WHERE product_id = $1 
        ORDER BY updated_at DESC
    `, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []Product
	for rows.Next() {
		var hist ProductHistory
		err := rows.Scan(
			&hist.HistoryID,
			&hist.ProductID,
			&hist.Label,
			&hist.Description,
			&hist.Price,
			&hist.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		product := Product{
			ID:          hist.ProductID,
			Label:       hist.Label,
			Description: hist.Description,
			Price:       hist.Price,
			CreatedAt:   hist.UpdatedAt,
		}
		history = append(history, product)
	}

	return history, nil
}
