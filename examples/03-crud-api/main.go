package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Yoochan45/go-api-utils/pkg/config"
	"github.com/Yoochan45/go-api-utils/pkg/database"
	"github.com/Yoochan45/go-api-utils/pkg/middleware"
	"github.com/Yoochan45/go-api-utils/pkg/repository"
	"github.com/Yoochan45/go-api-utils/pkg/request"
	"github.com/Yoochan45/go-api-utils/pkg/response"
)

// Product model
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

var db *sql.DB

func main() {
	// 1. Load config
	cfg := config.LoadEnv()

	// 2. Connect to database
	var err error
	db, err = database.ConnectPostgresURL(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// 3. Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/products", productsHandler)     // GET all, POST
	mux.HandleFunc("/products/", productByIDHandler) // GET by ID, PUT, DELETE

	// 4. Apply middleware
	handler := middleware.Logger(middleware.CORS(mux))

	// 5. Start server
	port := cfg.Port
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// Handler for /products (GET all & POST)
func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllProducts(w, r)
	case http.MethodPost:
		createProduct(w, r)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handler for /products/:id (GET by ID, PUT, DELETE)
func productByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getProductByID(w, r)
	case http.MethodPut:
		updateProduct(w, r)
	case http.MethodDelete:
		deleteProduct(w, r)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// GET /products - Get all products
func getAllProducts(w http.ResponseWriter, _ *http.Request) {
	query := repository.BuildSelectQuery("products",
		[]string{"id", "name", "description", "price", "stock"}, "")

	rows, err := db.Query(query)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch products")
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock); err != nil {
			response.InternalServerError(w, "Failed to scan products")
			return
		}
		products = append(products, p)
	}

	response.Success(w, "Products retrieved successfully", products)
}

// GET /products/:id - Get product by ID
func getProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := request.GetIDFromURL(r)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	query := repository.BuildSelectQuery("products",
		[]string{"id", "name", "description", "price", "stock"}, "id = $1")

	var p Product
	err = db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock)

	if err == sql.ErrNoRows {
		response.NotFound(w, "Product not found")
		return
	}
	if err != nil {
		response.InternalServerError(w, "Database error")
		return
	}

	response.Success(w, "Product retrieved successfully", p)
}

// POST /products - Create product
func createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := request.ParseJSON(r, &p); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if p.Name == "" || p.Price <= 0 {
		response.BadRequest(w, "Name and price are required")
		return
	}

	query := repository.BuildInsertQuery("products",
		[]string{"name", "description", "price", "stock"})

	err := db.QueryRow(query, p.Name, p.Description, p.Price, p.Stock).Scan(&p.ID)
	if err != nil {
		response.InternalServerError(w, "Failed to create product")
		return
	}

	response.Created(w, "Product created successfully", p)
}

// PUT /products/:id - Update product
func updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := request.GetIDFromURL(r)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	var p Product
	if err := request.ParseJSON(r, &p); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	query := repository.BuildUpdateQuery("products",
		[]string{"name", "description", "price", "stock"})

	result, err := db.Exec(query, p.Name, p.Description, p.Price, p.Stock, id)
	if err != nil {
		response.InternalServerError(w, "Failed to update product")
		return
	}

	if err := repository.CheckRowsAffected(result); err != nil {
		response.NotFound(w, "Product not found")
		return
	}

	p.ID = id
	response.Success(w, "Product updated successfully", p)
}

// DELETE /products/:id - Delete product
func deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := request.GetIDFromURL(r)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	result, err := db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		response.InternalServerError(w, "Failed to delete product")
		return
	}

	if err := repository.CheckRowsAffected(result); err != nil {
		response.NotFound(w, "Product not found")
		return
	}

	response.NoContent(w)
}
