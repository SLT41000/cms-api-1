package handler

import (
	"context"
	"log"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CustomerHandler struct {
	DB *pgx.Conn
}

// ListCustomers godoc
// @summary List Customers
// @description List all customers
// @tags customers
// @security ApiKeyAuth
// @id ListCustomers
// @accept json
// @produce json
// @response 200 {object} model.Customers "OK - Request successful"
// @response 201 {object} model.Customers "Created - Resource created successfully"
// @response 400 {object} model.Response "Bad Request - Invalid request parameters"
// @response 401 {object} model.Response "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.Response "Forbidden - Insufficient permissions"
// @response 404 {object} model.Response "Not Found - Resource doesn't exist"
// @response 422 {object} model.Response "Bad Request and Not Found (temporary)"
// @response 429 {object} model.Response "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.Response "Internal Server Error"
// @Router /api/v1/customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, `SELECT index,name FROM public."GoLangDemo"`)
	if err != nil {
		log.Printf("Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"data":    "",
		})
		return
	}
	defer rows.Close()

	var customers []model.CustomerOpt

	for rows.Next() {
		var cust model.CustomerOpt
		err := rows.Scan(&cust.Index, &cust.Name)
		if err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		customers = append(customers, cust)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All customers fetched",
		"data":    customers,
	})
}

// GetCustomer godoc
// @summary Get Customer
// @description  Get customer by id
// @tags customers
// @security ApiKeyAuth
// @id GetCustomer
// @accept json
// @produce json
// @param id path int true "id of customer to be gotten"
// @response 200 {object} model.Customers "OK - Request successful"
// @response 201 {object} model.Customers "Created - Resource created successfully"
// @response 400 {object} model.Response "Bad Request - Invalid request parameters"
// @response 401 {object} model.Response "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.Response "Forbidden - Insufficient permissions"
// @response 404 {object} model.Response "Not Found - Resource doesn't exist"
// @response 422 {object} model.Response "Bad Request and Not Found (temporary)"
// @response 429 {object} model.Response "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.Response "Internal Server Error"
// @Router /api/v1/customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")

	var cust model.CustomerOpt
	err := h.DB.QueryRow(ctx, `SELECT index,name FROM public."GoLangDemo" WHERE index = $1`, id).Scan(&cust.Index, &cust.Name)
	if err != nil {
		log.Printf("Query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"customer_id": id,
			"message":     err.Error(),
			"data":        "",
		})
		return
	}

	// Return a JSON response
	c.JSON(http.StatusOK, gin.H{
		"message":     "Customer get",
		"customer_id": id,
		"data":        cust,
	})
}

// CreateCustomer godoc
// @summary Create Customer
// @description Create new customer
// @tags customers
// @security ApiKeyAuth
// @id CreateCustomer
// @accept json
// @produce json
// @param Customer body model.CustomerForCreate true "Customer data to be created"
// @response 200 {object} model.Customers "OK - Request successful"
// @response 201 {object} model.Customers "Created - Resource created successfully"
// @response 400 {object} model.Response "Bad Request - Invalid request parameters"
// @response 401 {object} model.Response "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.Response "Forbidden - Insufficient permissions"
// @response 404 {object} model.Response "Not Found - Resource doesn't exist"
// @response 422 {object} model.Response "Bad Request and Not Found (temporary)"
// @response 429 {object} model.Response "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.Response "Internal Server Error"
// @Router /api/v1/customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req model.CustomerForCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var cust model.CustomerOpt
	query := `
		INSERT INTO public."GoLangDemo" (name)
		VALUES ($1)
		RETURNING index, name
	`

	err := h.DB.QueryRow(ctx, query, req.Firstname).
		Scan(&cust.Index, &cust.Name)
	if err != nil {
		log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed"})
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, gin.H{
		"message":   "Customer Insert",
		"Insert_id": cust.Index,
		"data":      cust,
	})
}

// DeleteCustomer deletes a customer by ID
// @summary Delete customer
// @description Deletes a customer by ID
// @id DeleteCustomer
// @tags customers
// @accept json
// @produce json
// @param id path int true "Customer ID"
// @response 200 {object} model.Customers "OK - Request successful"
// @response 201 {object} model.Customers "Created - Resource created successfully"
// @response 400 {object} model.Response "Bad Request - Invalid request parameters"
// @response 401 {object} model.Response "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.Response "Forbidden - Insufficient permissions"
// @response 404 {object} model.Response "Not Found - Resource doesn't exist"
// @response 422 {object} model.Response "Bad Request and Not Found (temporary)"
// @response 429 {object} model.Response "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.Response "Internal Server Error"
// @security ApiKeyAuth
// @router /api/v1/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")

	var cust model.CustomerOpt
	err := h.DB.QueryRow(ctx, `DELETE FROM public."GoLangDemo" WHERE index = $1 RETURNING index, name`, id).Scan(&cust.Index, &cust.Name)
	if err != nil {
		log.Printf("Delete failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"customer_id": id,
			"message":     err.Error(),
			"data":        "",
		})
		return
	}

	// Return a JSON response
	c.JSON(http.StatusOK, gin.H{
		"message":    "Customer deleted",
		"deleted_id": id,
		"data":       cust,
	})
}

// UpdateCustomer godoc
// @summary Update Customer
// @description Update customer by id
// @tags customers
// @security ApiKeyAuth
// @id UpdateCustomer
// @accept json
// @produce json
// @param id path int true "id of customer to be updated"
// @param Customer body model.CustomerForUpdate true "Customer data to be updated"
// @response 200 {object} model.Customers "OK - Request successful"
// @response 201 {object} model.Customers "Created - Resource created successfully"
// @response 400 {object} model.Response "Bad Request - Invalid request parameters"
// @response 401 {object} model.Response "Unauthorized - Invalid or missing authentication"
// @response 403 {object} model.Response "Forbidden - Insufficient permissions"
// @response 404 {object} model.Response "Not Found - Resource doesn't exist"
// @response 422 {object} model.Response "Bad Request and Not Found (temporary)"
// @response 429 {object} model.Response "Too Many Requests - Rate limit exceeded"
// @response 500 {object} model.Response "Internal Server Error"
// @Router /api/v1/customers/{id} [patch]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")

	var req model.CustomerForUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}
	if req.Firstname == nil { // enforce required field(s) as you like
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	newName := *req.Firstname

	var cust model.CustomerOpt
	err := h.DB.QueryRow(ctx, `
				UPDATE public."GoLangDemo"
				SET name = $2
				WHERE index = $1
				RETURNING index, name`, id, newName).Scan(&cust.Index, &cust.Name)
	if err != nil {
		log.Printf("Delete failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"customer_id": id,
			"message":     err.Error(),
			"data":        "",
		})
		return
	}

	// Return a JSON response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Customer Update",
		"update_id": id,
		"data":      cust,
	})
}
