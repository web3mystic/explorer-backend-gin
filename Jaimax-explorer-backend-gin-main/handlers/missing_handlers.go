package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ─── Address Summary ──────────────────────────────────────────────────────────

// GET /addresses/:address
// Returns a unified summary: balances, tx count, bank-transfer count,
// cw20-transfer count, and vote count — all in one call.
func GetAddressSummary(c *gin.Context) {
	address := c.Param("address")

	// balances
	balRows, err := config.DB.Query(`
		SELECT address, denom, amount, height, updated_at, created_at
		FROM balances WHERE address = $1 ORDER BY denom`, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer balRows.Close()
	balances := []models.Balance{}
	for balRows.Next() {
		var b models.Balance
		if err := balRows.Scan(&b.Address, &b.Denom, &b.Amount, &b.Height, &b.UpdatedAt, &b.CreatedAt); err == nil {
			balances = append(balances, b)
		}
	}

	// counts
	var txCount, bankCount, cw20Count, voteCount int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM address_transactions WHERE address = $1`, address).Scan(&txCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM bank_transfers WHERE from_address = $1 OR to_address = $1`, address).Scan(&bankCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM cw20_transfers WHERE from_address = $1 OR to_address = $1`, address).Scan(&cw20Count)
	config.DB.QueryRow(`SELECT COUNT(*) FROM votes WHERE voter = $1`, address).Scan(&voteCount)

	// authority status
	var isAuthority bool
	var authActive *bool
	config.DB.QueryRow(`SELECT active FROM authority_accounts WHERE address = $1`, address).Scan(&authActive)
	if authActive != nil {
		isAuthority = *authActive
	}

	c.JSON(http.StatusOK, gin.H{
		"address":            address,
		"balances":           balances,
		"tx_count":           txCount,
		"bank_transfer_count": bankCount,
		"cw20_transfer_count": cw20Count,
		"vote_count":         voteCount,
		"is_authority":       isAuthority,
	})
}

// ─── Validator Blocks ─────────────────────────────────────────────────────────

// GET /validators/:address/blocks
// Lists blocks proposed by a specific validator (uses proposer_address).
func GetValidatorBlocks(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM blocks WHERE proposer_address = $1`, address).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT height, hash, time, proposer_address, tx_count, created_at
		FROM blocks WHERE proposer_address = $1
		ORDER BY height DESC LIMIT $2 OFFSET $3`,
		address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	blocks := []models.Block{}
	for rows.Next() {
		var b models.Block
		if err := rows.Scan(&b.Height, &b.Hash, &b.Time, &b.ProposerAddress, &b.TxCount, &b.CreatedAt); err == nil {
			blocks = append(blocks, b)
		}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       blocks,
		Page:       pq.Page,
		PerPage:    pq.Limit(),
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// ─── Global WASM Events ───────────────────────────────────────────────────────

// GET /wasm/events
// Paginated list of all wasm events across all contracts.
func ListWasmEvents(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// optional filter by action type
	action := c.Query("action")

	var total int64
	if action != "" {
		config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_events WHERE action = $1`, action).Scan(&total)
	} else {
		config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_events`).Scan(&total)
	}

	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Close() error
	}
	var err error

	if action != "" {
		rows, err = config.DB.Query(`
			SELECT id, tx_hash, msg_index, event_index, height, contract_address, action, raw_attributes, timestamp, created_at
			FROM wasm_events WHERE action = $1
			ORDER BY height DESC LIMIT $2 OFFSET $3`,
			action, pq.Limit(), pq.Offset())
	} else {
		rows, err = config.DB.Query(`
			SELECT id, tx_hash, msg_index, event_index, height, contract_address, action, raw_attributes, timestamp, created_at
			FROM wasm_events
			ORDER BY height DESC LIMIT $1 OFFSET $2`,
			pq.Limit(), pq.Offset())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	events := []models.WasmEvent{}
	for rows.Next() {
		var e models.WasmEvent
		if err := rows.Scan(&e.ID, &e.TxHash, &e.MsgIndex, &e.EventIndex, &e.Height,
			&e.ContractAddress, &e.Action, &e.RawAttributes, &e.Timestamp, &e.CreatedAt); err == nil {
			events = append(events, e)
		}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       events,
		Page:       pq.Page,
		PerPage:    pq.Limit(),
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// ─── Single Wasm Execution ────────────────────────────────────────────────────

// GET /wasm/executions/:id
// Fetch a single wasm execution record by its primary-key id.
func GetWasmExecution(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution id"})
		return
	}

	var e models.WasmExecution
	err = config.DB.QueryRow(`
		SELECT id, tx_hash, msg_index, height, sender, contract_address,
		       execute_msg, execute_action, funds, gas_used, success, error, timestamp, created_at
		FROM wasm_executions WHERE id = $1`, id).
		Scan(&e.ID, &e.TxHash, &e.MsgIndex, &e.Height, &e.Sender, &e.ContractAddress,
			&e.ExecuteMsg, &e.ExecuteAction, &e.Funds, &e.GasUsed, &e.Success, &e.Error,
			&e.Timestamp, &e.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}
	c.JSON(http.StatusOK, e)
}

// ─── Instantiations by Code ───────────────────────────────────────────────────

// GET /wasm/codes/:code_id/instantiations
// Returns raw instantiation records (from wasm_instantiations) for a code id.
// Complements the existing /wasm/codes/:code_id/contracts endpoint which returns
// the live contract state from wasm_contracts.
func GetInstantiationsByCode(c *gin.Context) {
	codeID, err := strconv.ParseInt(c.Param("code_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code_id"})
		return
	}
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_instantiations WHERE code_id = $1`, codeID).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT id, tx_hash, msg_index, height, creator, admin, code_id, label,
		       contract_address, init_msg, funds, success, error, timestamp, created_at
		FROM wasm_instantiations WHERE code_id = $1
		ORDER BY height DESC LIMIT $2 OFFSET $3`,
		codeID, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	insts := []models.WasmInstantiation{}
	for rows.Next() {
		var i models.WasmInstantiation
		if err := rows.Scan(&i.ID, &i.TxHash, &i.MsgIndex, &i.Height, &i.Creator, &i.Admin,
			&i.CodeID, &i.Label, &i.ContractAddress, &i.InitMsg, &i.Funds,
			&i.Success, &i.Error, &i.Timestamp, &i.CreatedAt); err == nil {
			insts = append(insts, i)
		}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       insts,
		Page:       pq.Page,
		PerPage:    pq.Limit(),
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// ─── Balance Summary ──────────────────────────────────────────────────────────

// GET /balances/summary
// Returns per-denom stats: holder count and total supply (sum of amounts).
// Useful for a chain overview / token supply dashboard.
func GetBalancesSummary(c *gin.Context) {
	rows, err := config.DB.Query(`
		SELECT denom,
		       COUNT(DISTINCT address)   AS holders,
		       SUM(amount::numeric)      AS total_supply
		FROM balances
		GROUP BY denom
		ORDER BY total_supply DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type DenomSummary struct {
		Denom       string  `json:"denom"`
		Holders     int64   `json:"holders"`
		TotalSupply float64 `json:"total_supply"`
	}

	summary := []DenomSummary{}
	for rows.Next() {
		var d DenomSummary
		if err := rows.Scan(&d.Denom, &d.Holders, &d.TotalSupply); err == nil {
			summary = append(summary, d)
		}
	}

	c.JSON(http.StatusOK, gin.H{"denoms": summary, "count": len(summary)})
}