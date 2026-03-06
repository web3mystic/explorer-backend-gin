package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /transactions
func ListTransactions(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT hash, height, tx_index, gas_used, gas_wanted, fee, success, code, log, memo, msg_types, addresses, timestamp, created_at
		FROM transactions ORDER BY height DESC, tx_index LIMIT $1 OFFSET $2`,
		pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	txs := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		rows.Scan(&t.Hash, &t.Height, &t.TxIndex, &t.GasUsed, &t.GasWanted,
			&t.Fee, &t.Success, &t.Code, &t.Log, &t.Memo, &t.MsgTypes, &t.Addresses, &t.Timestamp, &t.CreatedAt)
		txs = append(txs, t)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: txs, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /transactions/:hash
func GetTransaction(c *gin.Context) {
	hash := c.Param("hash")
	var t models.Transaction
	err := config.DB.QueryRow(`
		SELECT hash, height, tx_index, gas_used, gas_wanted, fee, success, code, log, memo, msg_types, addresses, timestamp, created_at
		FROM transactions WHERE hash = $1`, hash).
		Scan(&t.Hash, &t.Height, &t.TxIndex, &t.GasUsed, &t.GasWanted,
			&t.Fee, &t.Success, &t.Code, &t.Log, &t.Memo, &t.MsgTypes, &t.Addresses, &t.Timestamp, &t.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// GET /transactions/recent
func ListRecentTransactions(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM recent_transactions`).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT hash, height, tx_index, success, gas_used, gas_wanted, fee, memo, timestamp, block_time
		FROM recent_transactions ORDER BY height DESC, tx_index LIMIT $1 OFFSET $2`,
		pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	txs := []models.RecentTransaction{}
	for rows.Next() {
		var t models.RecentTransaction
		rows.Scan(&t.Hash, &t.Height, &t.TxIndex, &t.Success, &t.GasUsed, &t.GasWanted, &t.Fee, &t.Memo, &t.Timestamp, &t.BlockTime)
		txs = append(txs, t)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: txs, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /transactions/:hash/messages
func GetTransactionMessages(c *gin.Context) {
	hash := c.Param("hash")
	rows, err := config.DB.Query(`
		SELECT id, tx_hash, msg_index, msg_type, sender, receiver, amount, denom, raw_data, created_at
		FROM messages WHERE tx_hash = $1 ORDER BY msg_index`, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	msgs := []models.Message{}
	for rows.Next() {
		var m models.Message
		rows.Scan(&m.ID, &m.TxHash, &m.MsgIndex, &m.MsgType, &m.Sender, &m.Receiver, &m.Amount, &m.Denom, &m.RawData, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	c.JSON(http.StatusOK, gin.H{"tx_hash": hash, "messages": msgs})
}

// GET /transactions/:hash/events
func GetTransactionEvents(c *gin.Context) {
	hash := c.Param("hash")
	rows, err := config.DB.Query(`
		SELECT id, tx_hash, event_index, event_type, attributes, created_at
		FROM events WHERE tx_hash = $1 ORDER BY event_index`, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		rows.Scan(&e.ID, &e.TxHash, &e.EventIndex, &e.EventType, &e.Attributes, &e.CreatedAt)
		events = append(events, e)
	}
	c.JSON(http.StatusOK, gin.H{"tx_hash": hash, "events": events})
}
