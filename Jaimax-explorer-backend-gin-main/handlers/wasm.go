package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ─── Wasm Codes ───────────────────────────────────────────────────────────────

// GET /wasm/codes
func ListWasmCodes(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_codes`).Scan(&total)
	rows, err := config.DB.Query(`SELECT code_id, creator, checksum, permission, uploaded_height, uploaded_time, upload_tx_hash, created_at FROM wasm_codes ORDER BY code_id DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	codes := []models.WasmCode{}
	for rows.Next() {
		var wc models.WasmCode
		rows.Scan(&wc.CodeID, &wc.Creator, &wc.Checksum, &wc.Permission, &wc.UploadedHeight, &wc.UploadedTime, &wc.UploadTxHash, &wc.CreatedAt)
		codes = append(codes, wc)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: codes, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /wasm/codes/:code_id
func GetWasmCode(c *gin.Context) {
	codeID, err := strconv.ParseInt(c.Param("code_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code_id"})
		return
	}
	var wc models.WasmCode
	err = config.DB.QueryRow(`SELECT code_id, creator, checksum, permission, uploaded_height, uploaded_time, upload_tx_hash, created_at FROM wasm_codes WHERE code_id = $1`, codeID).
		Scan(&wc.CodeID, &wc.Creator, &wc.Checksum, &wc.Permission, &wc.UploadedHeight, &wc.UploadedTime, &wc.UploadTxHash, &wc.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "code not found"})
		return
	}
	c.JSON(http.StatusOK, wc)
}

// ─── Wasm Contracts ───────────────────────────────────────────────────────────

const contractCols = `contract_address, code_id, creator, admin, label, init_msg, contract_info, instantiated_at_height, instantiated_at_time, instantiate_tx_hash, current_code_id, last_migrated_height, last_migrated_tx_hash, is_active, updated_at, created_at`

func scanContract(s interface{ Scan(...interface{}) error }) models.WasmContract {
	var wc models.WasmContract
	s.Scan(&wc.ContractAddress, &wc.CodeID, &wc.Creator, &wc.Admin, &wc.Label, &wc.InitMsg, &wc.ContractInfo,
		&wc.InstantiatedAtHeight, &wc.InstantiatedAtTime, &wc.InstantiateTxHash, &wc.CurrentCodeID,
		&wc.LastMigratedHeight, &wc.LastMigratedTxHash, &wc.IsActive, &wc.UpdatedAt, &wc.CreatedAt)
	return wc
}

// GET /wasm/contracts
func ListWasmContracts(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_contracts`).Scan(&total)
	rows, err := config.DB.Query(`SELECT `+contractCols+` FROM wasm_contracts ORDER BY instantiated_at_height DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	contracts := []models.WasmContract{}
	for rows.Next() {
		contracts = append(contracts, scanContract(rows))
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: contracts, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /wasm/contracts/:address
func GetWasmContract(c *gin.Context) {
	address := c.Param("address")
	row := config.DB.QueryRow(`SELECT `+contractCols+` FROM wasm_contracts WHERE contract_address = $1`, address)
	wc := scanContract(row)
	if wc.ContractAddress == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	c.JSON(http.StatusOK, wc)
}

// GET /wasm/codes/:code_id/contracts
func GetContractsByCode(c *gin.Context) {
	codeID, err := strconv.ParseInt(c.Param("code_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code_id"})
		return
	}
	rows, err := config.DB.Query(`SELECT `+contractCols+` FROM wasm_contracts WHERE code_id = $1 ORDER BY instantiated_at_height DESC`, codeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	contracts := []models.WasmContract{}
	for rows.Next() {
		contracts = append(contracts, scanContract(rows))
	}
	c.JSON(http.StatusOK, gin.H{"code_id": codeID, "contracts": contracts})
}

// ─── Wasm Executions ──────────────────────────────────────────────────────────

// GET /wasm/executions
func ListWasmExecutions(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_executions`).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, sender, contract_address, execute_msg, execute_action, funds, gas_used, success, error, timestamp, created_at FROM wasm_executions ORDER BY height DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	execs := []models.WasmExecution{}
	for rows.Next() {
		var e models.WasmExecution
		rows.Scan(&e.ID, &e.TxHash, &e.MsgIndex, &e.Height, &e.Sender, &e.ContractAddress, &e.ExecuteMsg, &e.ExecuteAction, &e.Funds, &e.GasUsed, &e.Success, &e.Error, &e.Timestamp, &e.CreatedAt)
		execs = append(execs, e)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: execs, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /wasm/contracts/:address/executions
func GetContractExecutions(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_executions WHERE contract_address = $1`, address).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, sender, contract_address, execute_msg, execute_action, funds, gas_used, success, error, timestamp, created_at FROM wasm_executions WHERE contract_address = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`, address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	execs := []models.WasmExecution{}
	for rows.Next() {
		var e models.WasmExecution
		rows.Scan(&e.ID, &e.TxHash, &e.MsgIndex, &e.Height, &e.Sender, &e.ContractAddress, &e.ExecuteMsg, &e.ExecuteAction, &e.Funds, &e.GasUsed, &e.Success, &e.Error, &e.Timestamp, &e.CreatedAt)
		execs = append(execs, e)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: execs, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// ─── Wasm Events ──────────────────────────────────────────────────────────────

// GET /wasm/contracts/:address/events
func GetContractEvents(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_events WHERE contract_address = $1`, address).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, event_index, height, contract_address, action, raw_attributes, timestamp, created_at FROM wasm_events WHERE contract_address = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`, address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	events := []models.WasmEvent{}
	for rows.Next() {
		var e models.WasmEvent
		rows.Scan(&e.ID, &e.TxHash, &e.MsgIndex, &e.EventIndex, &e.Height, &e.ContractAddress, &e.Action, &e.RawAttributes, &e.Timestamp, &e.CreatedAt)
		events = append(events, e)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: events, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// ─── Wasm Instantiations ──────────────────────────────────────────────────────

// GET /wasm/instantiations
func ListWasmInstantiations(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_instantiations`).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, creator, admin, code_id, label, contract_address, init_msg, funds, success, error, timestamp, created_at FROM wasm_instantiations ORDER BY height DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	insts := []models.WasmInstantiation{}
	for rows.Next() {
		var i models.WasmInstantiation
		rows.Scan(&i.ID, &i.TxHash, &i.MsgIndex, &i.Height, &i.Creator, &i.Admin, &i.CodeID, &i.Label, &i.ContractAddress, &i.InitMsg, &i.Funds, &i.Success, &i.Error, &i.Timestamp, &i.CreatedAt)
		insts = append(insts, i)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: insts, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// ─── Wasm Migrations ──────────────────────────────────────────────────────────

// GET /wasm/contracts/:address/migrations
func GetContractMigrations(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_migrations WHERE contract_address = $1`, address).Scan(&total)
	rows, err := config.DB.Query(`SELECT id, tx_hash, msg_index, height, sender, contract_address, old_code_id, new_code_id, migrate_msg, success, error, timestamp, created_at FROM wasm_migrations WHERE contract_address = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`, address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	migrations := []models.WasmMigration{}
	for rows.Next() {
		var m models.WasmMigration
		rows.Scan(&m.ID, &m.TxHash, &m.MsgIndex, &m.Height, &m.Sender, &m.ContractAddress, &m.OldCodeID, &m.NewCodeID, &m.MigrateMsg, &m.Success, &m.Error, &m.Timestamp, &m.CreatedAt)
		migrations = append(migrations, m)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: migrations, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// ─── Contract Activity ────────────────────────────────────────────────────────

// GET /wasm/activity
func ListContractActivity(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM contract_activity`).Scan(&total)
	rows, err := config.DB.Query(`SELECT contract_address, label, code_id, creator, total_executions, last_execution, unique_users FROM contract_activity ORDER BY total_executions DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	activities := []models.ContractActivity{}
	for rows.Next() {
		var a models.ContractActivity
		rows.Scan(&a.ContractAddress, &a.Label, &a.CodeID, &a.Creator, &a.TotalExecutions, &a.LastExecution, &a.UniqueUsers)
		activities = append(activities, a)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: activities, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}

// GET /wasm/activity/recent
func ListRecentWasmActivity(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM recent_wasm_activity`).Scan(&total)
	rows, err := config.DB.Query(`SELECT tx_hash, height, contract_address, contract_label, action, sender, recipient, amount, timestamp FROM recent_wasm_activity ORDER BY height DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	activities := []models.RecentWasmActivity{}
	for rows.Next() {
		var a models.RecentWasmActivity
		rows.Scan(&a.TxHash, &a.Height, &a.ContractAddress, &a.ContractLabel, &a.Action, &a.Sender, &a.Recipient, &a.Amount, &a.Timestamp)
		activities = append(activities, a)
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Data: activities, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
}
