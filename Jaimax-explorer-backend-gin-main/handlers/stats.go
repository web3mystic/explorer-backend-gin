package handlers

import (
	"blockchain-api/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /stats
func GetStats(c *gin.Context) {
	stats := gin.H{}

	var blockCount, txCount, validatorCount, proposalCount, wasmCodeCount, wasmContractCount int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM blocks`).Scan(&blockCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&txCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM validators`).Scan(&validatorCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM proposals`).Scan(&proposalCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_codes`).Scan(&wasmCodeCount)
	config.DB.QueryRow(`SELECT COUNT(*) FROM wasm_contracts`).Scan(&wasmContractCount)

	var latestHeight int64
	var latestHash, latestProposer string
	config.DB.QueryRow(`SELECT height, hash, proposer_address FROM blocks ORDER BY height DESC LIMIT 1`).
		Scan(&latestHeight, &latestHash, &latestProposer)

	stats["total_blocks"] = blockCount
	stats["total_transactions"] = txCount
	stats["total_validators"] = validatorCount
	stats["total_proposals"] = proposalCount
	stats["total_wasm_codes"] = wasmCodeCount
	stats["total_wasm_contracts"] = wasmContractCount
	stats["latest_block_height"] = latestHeight
	stats["latest_block_hash"] = latestHash
	stats["latest_block_proposer"] = latestProposer

	c.JSON(http.StatusOK, stats)
}

// GET /search?q=<hash | address | height | proposal_id | code_id | moniker>
//
// Searches across: blocks (hash or height), transactions, validators
// (operator/consensus address or moniker), wasm contracts, wasm codes,
// proposals (id or title prefix), and addresses.
// Returns all matching types so the frontend can route the user to the right page.
func Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	found := []map[string]interface{}{}

	// ── Block by hash or height ──────────────────────────────────────────────
	var blockHeight int64
	if err := config.DB.QueryRow(
		`SELECT height FROM blocks WHERE hash = $1 OR CAST(height AS text) = $1 LIMIT 1`, q,
	).Scan(&blockHeight); err == nil {
		found = append(found, map[string]interface{}{"type": "block", "value": blockHeight})
	}

	// ── Transaction by hash ──────────────────────────────────────────────────
	var txHash string
	if err := config.DB.QueryRow(
		`SELECT hash FROM transactions WHERE hash = $1 LIMIT 1`, q,
	).Scan(&txHash); err == nil {
		found = append(found, map[string]interface{}{"type": "transaction", "value": txHash})
	}

	// ── Validator by operator address, consensus address, or moniker ─────────
	var opAddress, moniker string
	if err := config.DB.QueryRow(
		`SELECT operator_address, moniker FROM validators
		 WHERE operator_address = $1 OR consensus_address = $1 OR moniker ILIKE $2
		 LIMIT 1`, q, "%"+q+"%",
	).Scan(&opAddress, &moniker); err == nil {
		found = append(found, map[string]interface{}{
			"type":    "validator",
			"value":   opAddress,
			"moniker": moniker,
		})
	}

	// ── Wasm contract by address or label ────────────────────────────────────
	var contractAddress, contractLabel string
	if err := config.DB.QueryRow(
		`SELECT contract_address, label FROM wasm_contracts
		 WHERE contract_address = $1 OR label ILIKE $2
		 LIMIT 1`, q, "%"+q+"%",
	).Scan(&contractAddress, &contractLabel); err == nil {
		found = append(found, map[string]interface{}{
			"type":  "contract",
			"value": contractAddress,
			"label": contractLabel,
		})
	}

	// ── Wasm code by code_id ─────────────────────────────────────────────────
	var codeID int64
	if err := config.DB.QueryRow(
		`SELECT code_id FROM wasm_codes WHERE CAST(code_id AS text) = $1 LIMIT 1`, q,
	).Scan(&codeID); err == nil {
		found = append(found, map[string]interface{}{"type": "wasm_code", "value": codeID})
	}

	// ── Proposal by id or title prefix ───────────────────────────────────────
	var proposalID int64
	var proposalTitle, proposalStatus string
	if err := config.DB.QueryRow(
		`SELECT proposal_id, title, status FROM proposals
		 WHERE CAST(proposal_id AS text) = $1 OR title ILIKE $2
		 ORDER BY proposal_id DESC LIMIT 1`, q, "%"+q+"%",
	).Scan(&proposalID, &proposalTitle, &proposalStatus); err == nil {
		found = append(found, map[string]interface{}{
			"type":   "proposal",
			"value":  proposalID,
			"title":  proposalTitle,
			"status": proposalStatus,
		})
	}

	// ── Address (has transactions or balances) ───────────────────────────────
	var addr string
	if err := config.DB.QueryRow(
		`SELECT address FROM address_transactions WHERE address = $1 LIMIT 1`, q,
	).Scan(&addr); err == nil {
		found = append(found, map[string]interface{}{"type": "address", "value": addr})
	} else {
		// fallback: check balances table in case address has balance but no indexed tx
		if err2 := config.DB.QueryRow(
			`SELECT address FROM balances WHERE address = $1 LIMIT 1`, q,
		).Scan(&addr); err2 == nil {
			found = append(found, map[string]interface{}{"type": "address", "value": addr})
		}
	}

	if len(found) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no results found", "query": q})
		return
	}

	c.JSON(http.StatusOK, gin.H{"query": q, "results": found})
}