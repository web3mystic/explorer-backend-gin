package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TxWithTransfers is a Transaction enriched with its bank_transfer legs.
// Declared at package level so both GetBlockTransactions and totalTransfers can use it.
type TxWithTransfers struct {
	models.Transaction
	Transfers []models.BankTransfer `json:"transfers"`
}
// @Summary      List Blocks
// @Tags         blocks
// @Produce      json
// @Param        page      query  int  false  "Page"
// @Param        per_page  query  int  false  "Per page"
// @Success      200  {object}  models.PaginatedResponse
// @Router       /blocks [get]
// GET /blocks
func ListBlocks(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM blocks`).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT height, hash, time, proposer_address, tx_count, created_at
		FROM blocks ORDER BY height DESC LIMIT $1 OFFSET $2`,
		pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	blocks := []models.Block{}
	for rows.Next() {
		var b models.Block
		rows.Scan(&b.Height, &b.Hash, &b.Time, &b.ProposerAddress, &b.TxCount, &b.CreatedAt)
		blocks = append(blocks, b)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       blocks,
		Page:       pq.Page,
		PerPage:    pq.Limit(),
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}
// @Summary      Get latest block
// @Description  Returns the most recently indexed block
// @Tags         Blocks
// @Produce      json
// @Success      200  {object}  models.Block
// @Failure      404  {object}  map[string]string
// @Router       /blocks/latest [get]
// GET /blocks/latest
func GetLatestBlock(c *gin.Context) {
	var b models.Block
	err := config.DB.QueryRow(`
		SELECT height, hash, time, proposer_address, tx_count, created_at
		FROM blocks ORDER BY height DESC LIMIT 1`).
		Scan(&b.Height, &b.Hash, &b.Time, &b.ProposerAddress, &b.TxCount, &b.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no blocks found"})
		return
	}
	c.JSON(http.StatusOK, b)
}
// @Summary      Get block by height
// @Description  Returns a single block by its height
// @Tags         Blocks
// @Produce      json
// @Param        height  path  int  true  "Block height"
// @Success      200  {object}  models.Block
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /blocks/{height} [get]
// GET /blocks/:height
func GetBlock(c *gin.Context) {
	height, err := strconv.ParseInt(c.Param("height"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid height"})
		return
	}

	var b models.Block
	err = config.DB.QueryRow(`
		SELECT height, hash, time, proposer_address, tx_count, created_at
		FROM blocks WHERE height = $1`, height).
		Scan(&b.Height, &b.Hash, &b.Time, &b.ProposerAddress, &b.TxCount, &b.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
		return
	}
	c.JSON(http.StatusOK, b)
}

// GET /blocks/:height/transactions
//
// Returns every tx in the block. Each tx carries a "transfers" array from
// bank_transfers — one element per from->to leg, so a MsgMultiSend to 5
// recipients shows all 5 rows instead of just 1.

// @Summary      Get transactions in a block
// @Description  Returns all transactions in a block enriched with their bank transfer legs. MsgMultiSend with 5 outputs shows all 5 transfer rows.
// @Tags         Blocks
// @Produce      json
// @Param        height  path  int  true  "Block height"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blocks/{height}/transactions [get]
func GetBlockTransactions(c *gin.Context) {
	height, err := strconv.ParseInt(c.Param("height"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid height"})
		return
	}

	// 1. fetch all transactions for this block
	txRows, err := config.DB.Query(`
		SELECT hash, height, tx_index, gas_used, gas_wanted,
		       fee, success, code, log, memo,
		       msg_types, addresses, timestamp, created_at
		FROM transactions
		WHERE height = $1
		ORDER BY tx_index`, height)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer txRows.Close()

	txMap := map[string]*TxWithTransfers{}
	txOrder := []string{}

	for txRows.Next() {
		var t models.Transaction
		txRows.Scan(
			&t.Hash, &t.Height, &t.TxIndex, &t.GasUsed, &t.GasWanted,
			&t.Fee, &t.Success, &t.Code, &t.Log, &t.Memo,
			&t.MsgTypes, &t.Addresses, &t.Timestamp, &t.CreatedAt,
		)
		txMap[t.Hash] = &TxWithTransfers{
			Transaction: t,
			Transfers:   []models.BankTransfer{},
		}
		txOrder = append(txOrder, t.Hash)
	}

	// 2. fetch every bank_transfer row for this block
	// bank_transfers has one row per individual send leg.
	// MsgMultiSend with 5 outputs = 5 rows, same tx_hash, different msg_index.
	btRows, err := config.DB.Query(`
		SELECT id, tx_hash, msg_index, height,
		       from_address, to_address, amount, denom,
		       amount_value, timestamp, created_at
		FROM bank_transfers
		WHERE height = $1
		ORDER BY tx_hash, msg_index`, height)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer btRows.Close()

	for btRows.Next() {
		var bt models.BankTransfer
		btRows.Scan(
			&bt.ID, &bt.TxHash, &bt.MsgIndex, &bt.Height,
			&bt.FromAddress, &bt.ToAddress, &bt.Amount, &bt.Denom,
			&bt.AmountValue, &bt.Timestamp, &bt.CreatedAt,
		)
		if tx, ok := txMap[bt.TxHash]; ok {
			tx.Transfers = append(tx.Transfers, bt)
		}
	}

	// 3. build ordered result slice
	result := make([]TxWithTransfers, 0, len(txOrder))
	for _, hash := range txOrder {
		result = append(result, *txMap[hash])
	}

	c.JSON(http.StatusOK, gin.H{
		"height":          height,
		"tx_count":        len(result),
		"total_transfers": totalTransfers(result),
		"transactions":    result,
	})
}

// GET /blocks/:height/transfers
//
// Flat list of every bank_transfer in the block — all money movements
// without grouping by tx.
// @Summary      Get transfers in a block
// @Description  Returns a flat list of all bank transfers in a block (all money movements without grouping by tx)
// @Tags         Blocks
// @Produce      json
// @Param        height  path  int  true  "Block height"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blocks/{height}/transfers [get]
func GetBlockTransfers(c *gin.Context) {
	height, err := strconv.ParseInt(c.Param("height"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid height"})
		return
	}

	rows, err := config.DB.Query(`
		SELECT id, tx_hash, msg_index, height,
		       from_address, to_address, amount, denom,
		       amount_value, timestamp, created_at
		FROM bank_transfers
		WHERE height = $1
		ORDER BY tx_hash, msg_index`, height)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	transfers := []models.BankTransfer{}
	for rows.Next() {
		var bt models.BankTransfer
		rows.Scan(
			&bt.ID, &bt.TxHash, &bt.MsgIndex, &bt.Height,
			&bt.FromAddress, &bt.ToAddress, &bt.Amount, &bt.Denom,
			&bt.AmountValue, &bt.Timestamp, &bt.CreatedAt,
		)
		transfers = append(transfers, bt)
	}

	c.JSON(http.StatusOK, gin.H{
		"height":    height,
		"count":     len(transfers),
		"transfers": transfers,
	})
}

// totalTransfers counts all transfer legs across a slice of enriched txs.
func totalTransfers(txs []TxWithTransfers) int {
	n := 0
	for _, t := range txs {
		n += len(t.Transfers)
	}
	return n
}