package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /addresses/:address/balances
func GetAddressBalances(c *gin.Context) {
	address := c.Param("address")
	rows, err := config.DB.Query(`
		SELECT address, denom, amount, height, updated_at, created_at
		FROM balances WHERE address = $1 ORDER BY denom`, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	balances := []models.Balance{}
	for rows.Next() {
		var b models.Balance
		if err := rows.Scan(&b.Address, &b.Denom, &b.Amount, &b.Height, &b.UpdatedAt, &b.CreatedAt); err == nil {
			balances = append(balances, b)
		}
	}
	c.JSON(http.StatusOK, gin.H{"address": address, "balances": balances})
}

// GET /addresses/:address/transactions
func GetAddressTransactions(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM address_transactions WHERE address = $1`, address).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT address, tx_hash, height, created_at
		FROM address_transactions WHERE address = $1
		ORDER BY height DESC LIMIT $2 OFFSET $3`,
		address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	txs := []models.AddressTransaction{}
	for rows.Next() {
		var t models.AddressTransaction
		if err := rows.Scan(&t.Address, &t.TxHash, &t.Height, &t.CreatedAt); err == nil {
			txs = append(txs, t)
		}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       txs,
		Page:       pq.Page,
		PerPage:    pq.Limit(),
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /balances
// Lists all balances paginated. Optional ?denom= filter.
func ListBalances(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	denom := c.Query("denom")

	var total int64
	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Close() error
	}
	var err error

	if denom != "" {
		config.DB.QueryRow(`SELECT COUNT(*) FROM balances WHERE denom = $1`, denom).Scan(&total)
		rows, err = config.DB.Query(
			`SELECT address, denom, amount, height, updated_at, created_at
			 FROM balances WHERE denom = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`,
			denom, pq.Limit(), pq.Offset())
	} else {
		config.DB.QueryRow(`SELECT COUNT(*) FROM balances`).Scan(&total)
		rows, err = config.DB.Query(
			`SELECT address, denom, amount, height, updated_at, created_at
			 FROM balances ORDER BY height DESC LIMIT $1 OFFSET $2`,
			pq.Limit(), pq.Offset())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	balances := []models.Balance{}
	for rows.Next() {
		var b models.Balance
		if err := rows.Scan(&b.Address, &b.Denom, &b.Amount, &b.Height, &b.UpdatedAt, &b.CreatedAt); err == nil {
			balances = append(balances, b)
		}
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       balances,
		Page:       pq.Page,
		PerPage:    pq.Limit(),
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}