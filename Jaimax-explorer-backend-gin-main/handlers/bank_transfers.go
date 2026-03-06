package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

func scanBankTransfer(rows interface{ Scan(...interface{}) error }) (models.BankTransfer, error) {
	var t models.BankTransfer
	err := rows.Scan(&t.ID, &t.TxHash, &t.MsgIndex, &t.Height, &t.FromAddress,
		&t.ToAddress, &t.Amount, &t.Denom, &t.AmountValue, &t.Timestamp, &t.CreatedAt)
	return t, err
}

const bankTransferCols = `id, tx_hash, msg_index, height, from_address, to_address, amount, denom, amount_value, timestamp, created_at`

// GET /bank-transfers
func ListBankTransfers(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM bank_transfers`).Scan(&total)

	rows, err := config.DB.Query(`SELECT `+bankTransferCols+` FROM bank_transfers ORDER BY height DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	transfers := []models.BankTransfer{}
	for rows.Next() {
		t, _ := scanBankTransfer(rows)
		transfers = append(transfers, t)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: transfers, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /bank-transfers/:id
func GetBankTransfer(c *gin.Context) {
	id := c.Param("id")
	var t models.BankTransfer
	err := config.DB.QueryRow(`SELECT `+bankTransferCols+` FROM bank_transfers WHERE id = $1`, id).
		Scan(&t.ID, &t.TxHash, &t.MsgIndex, &t.Height, &t.FromAddress,
			&t.ToAddress, &t.Amount, &t.Denom, &t.AmountValue, &t.Timestamp, &t.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// GET /addresses/:address/bank-transfers
func GetAddressBankTransfers(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM bank_transfers WHERE from_address = $1 OR to_address = $1`, address).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT `+bankTransferCols+` FROM bank_transfers
		WHERE from_address = $1 OR to_address = $1
		ORDER BY height DESC LIMIT $2 OFFSET $3`, address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	transfers := []models.BankTransfer{}
	for rows.Next() {
		t, _ := scanBankTransfer(rows)
		transfers = append(transfers, t)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: transfers, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}
