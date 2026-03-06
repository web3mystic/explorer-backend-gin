package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─── CW20 Transfers ───────────────────────────────────────────────────────────

// GET /cw20/transfers
func ListCW20Transfers(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contract := c.Query("contract")

	var total int64
	if contract != "" {
		config.DB.QueryRow(`SELECT COUNT(*) FROM cw20_transfers WHERE contract_address = $1`, contract).Scan(&total)
		rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, contract_address, action, from_address, to_address, amount, memo, raw_attributes, timestamp, created_at FROM cw20_transfers WHERE contract_address = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`, contract, pq.Limit(), pq.Offset())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		c.JSON(http.StatusOK, models.PaginatedResponse{Data: scanCW20Transfers(rows), Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
		return
	}

	config.DB.QueryRow(`SELECT COUNT(*) FROM cw20_transfers`).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, contract_address, action, from_address, to_address, amount, memo, raw_attributes, timestamp, created_at FROM cw20_transfers ORDER BY height DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: scanCW20Transfers(rows), Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /addresses/:address/cw20-transfers
func GetAddressCW20Transfers(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM cw20_transfers WHERE from_address = $1 OR to_address = $1`, address).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, contract_address, action, from_address, to_address, amount, memo, raw_attributes, timestamp, created_at FROM cw20_transfers WHERE from_address = $1 OR to_address = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`, address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: scanCW20Transfers(rows), Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /cw20/address-activity
func ListCW20AddressActivity(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contract := c.Query("contract")

	var total int64
	baseQ := `SELECT contract_address, address, transfers_sent, total_sent FROM cw20_address_activity`
	countQ := `SELECT COUNT(*) FROM cw20_address_activity`

	if contract != "" {
		config.DB.QueryRow(countQ+` WHERE contract_address = $1`, contract).Scan(&total)
		rows, err := config.DB.Query(baseQ+` WHERE contract_address = $1 ORDER BY total_sent DESC LIMIT $2 OFFSET $3`, contract, pq.Limit(), pq.Offset())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		c.JSON(http.StatusOK, models.PaginatedResponse{Data: scanCW20Activity(rows), Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
		return
	}

	config.DB.QueryRow(countQ).Scan(&total)
	rows, err := config.DB.Query(baseQ+` ORDER BY total_sent DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: scanCW20Activity(rows), Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

func scanCW20Transfers(rows interface {
	Next() bool
	Scan(...interface{}) error
}) []models.CW20Transfer {
	transfers := []models.CW20Transfer{}
	for rows.Next() {
		var t models.CW20Transfer
		rows.Scan(&t.ID, &t.TxHash, &t.MsgIndex, &t.Height, &t.ContractAddress, &t.Action, &t.FromAddress, &t.ToAddress, &t.Amount, &t.Memo, &t.RawAttributes, &t.Timestamp, &t.CreatedAt)
		transfers = append(transfers, t)
	}
	return transfers
}

func scanCW20Activity(rows interface {
	Next() bool
	Scan(...interface{}) error
}) []models.CW20AddressActivity {
	activities := []models.CW20AddressActivity{}
	for rows.Next() {
		var a models.CW20AddressActivity
		rows.Scan(&a.ContractAddress, &a.Address, &a.TransfersSent, &a.TotalSent)
		activities = append(activities, a)
	}
	return activities
}

// ─── Authority Accounts ───────────────────────────────────────────────────────

// GET /authority-accounts
func ListAuthorityAccounts(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	active := c.Query("active")

	var total int64
	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Close() error
	}
	var err error

	cols := `address, added_at, added_by, added_at_height, active, removed_at, removed_by, source, updated_at, created_at`
	if active != "" {
		isActive := active == "true"
		config.DB.QueryRow(`SELECT COUNT(*) FROM authority_accounts WHERE active = $1`, isActive).Scan(&total)
		rows, err = config.DB.Query(`SELECT `+cols+` FROM authority_accounts WHERE active = $1 ORDER BY added_at DESC LIMIT $2 OFFSET $3`, isActive, pq.Limit(), pq.Offset())
	} else {
		config.DB.QueryRow(`SELECT COUNT(*) FROM authority_accounts`).Scan(&total)
		rows, err = config.DB.Query(`SELECT `+cols+` FROM authority_accounts ORDER BY added_at DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	accounts := []models.AuthorityAccount{}
	for rows.Next() {
		var a models.AuthorityAccount
		rows.Scan(&a.Address, &a.AddedAt, &a.AddedBy, &a.AddedAtHeight, &a.Active, &a.RemovedAt, &a.RemovedBy, &a.Source, &a.UpdatedAt, &a.CreatedAt)
		accounts = append(accounts, a)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: accounts, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /authority-accounts/:address
func GetAuthorityAccount(c *gin.Context) {
	address := c.Param("address")
	var a models.AuthorityAccount
	err := config.DB.QueryRow(`SELECT address, added_at, added_by, added_at_height, active, removed_at, removed_by, source, updated_at, created_at FROM authority_accounts WHERE address = $1`, address).
		Scan(&a.Address, &a.AddedAt, &a.AddedBy, &a.AddedAtHeight, &a.Active, &a.RemovedAt, &a.RemovedBy, &a.Source, &a.UpdatedAt, &a.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

// ─── Sync / Indexer State ─────────────────────────────────────────────────────

// GET /state/indexer
func GetIndexerState(c *gin.Context) {
	var s models.IndexerState
	err := config.DB.QueryRow(`SELECT id, last_height, last_block_hash, updated_at FROM indexer_state ORDER BY id LIMIT 1`).
		Scan(&s.ID, &s.LastHeight, &s.LastBlockHash, &s.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "indexer state not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GET /state/validator-sync
func GetValidatorSyncState(c *gin.Context) {
	var s models.ValidatorSyncState
	err := config.DB.QueryRow(`SELECT id, last_sync_height, last_sync_time, updated_at FROM validator_sync_state ORDER BY id LIMIT 1`).
		Scan(&s.ID, &s.LastSyncHeight, &s.LastSyncTime, &s.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "validator sync state not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GET /state/proposal-sync
func GetProposalSyncState(c *gin.Context) {
	var s models.ProposalSyncState
	err := config.DB.QueryRow(`SELECT id, last_sync_height, last_sync_time, updated_at FROM proposal_sync_state ORDER BY id LIMIT 1`).
		Scan(&s.ID, &s.LastSyncHeight, &s.LastSyncTime, &s.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proposal sync state not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}
