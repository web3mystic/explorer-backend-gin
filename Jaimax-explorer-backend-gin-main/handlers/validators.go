package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /validators
func ListValidators(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status := c.Query("status") // optional filter: BOND_STATUS_BONDED, etc.

	baseQuery := `FROM validators`
	args := []interface{}{}
	where := ``
	if status != "" {
		where = ` WHERE status = $1`
		args = append(args, status)
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) `+baseQuery+where, args...).Scan(&total)

	limitOffset := []interface{}{pq.Limit(), pq.Offset()}
	if len(args) > 0 {
		limitOffset = append(args, pq.Limit(), pq.Offset())
		where += ` ORDER BY voting_power DESC LIMIT $2 OFFSET $3`
	} else {
		where += ` ORDER BY voting_power DESC LIMIT $1 OFFSET $2`
	}

	rows, err := config.DB.Query(`SELECT operator_address, consensus_address, consensus_pubkey,
		moniker, identity, website, security_contact, details, jailed, status, power, voting_power,
		tokens, delegator_shares, added_by, added_at, added_at_height, unbonding_height, unbonding_time,
		commission_rate, commission_max_rate, commission_max_change_rate, min_self_delegation, updated_at, created_at
		`+baseQuery+where, limitOffset...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	validators := []models.Validator{}
	for rows.Next() {
		var v models.Validator
		rows.Scan(&v.OperatorAddress, &v.ConsensusAddress, &v.ConsensusPubkey, &v.Moniker,
			&v.Identity, &v.Website, &v.SecurityContact, &v.Details, &v.Jailed, &v.Status,
			&v.Power, &v.VotingPower, &v.Tokens, &v.DelegatorShares, &v.AddedBy, &v.AddedAt,
			&v.AddedAtHeight, &v.UnbondingHeight, &v.UnbondingTime, &v.CommissionRate,
			&v.CommissionMaxRate, &v.CommissionMaxChangeRate, &v.MinSelfDelegation, &v.UpdatedAt, &v.CreatedAt)
		validators = append(validators, v)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: validators, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /validators/:address
func GetValidator(c *gin.Context) {
	address := c.Param("address")
	var v models.Validator
	err := config.DB.QueryRow(`SELECT operator_address, consensus_address, consensus_pubkey,
		moniker, identity, website, security_contact, details, jailed, status, power, voting_power,
		tokens, delegator_shares, added_by, added_at, added_at_height, unbonding_height, unbonding_time,
		commission_rate, commission_max_rate, commission_max_change_rate, min_self_delegation, updated_at, created_at
		FROM validators WHERE operator_address = $1`, address).
		Scan(&v.OperatorAddress, &v.ConsensusAddress, &v.ConsensusPubkey, &v.Moniker,
			&v.Identity, &v.Website, &v.SecurityContact, &v.Details, &v.Jailed, &v.Status,
			&v.Power, &v.VotingPower, &v.Tokens, &v.DelegatorShares, &v.AddedBy, &v.AddedAt,
			&v.AddedAtHeight, &v.UnbondingHeight, &v.UnbondingTime, &v.CommissionRate,
			&v.CommissionMaxRate, &v.CommissionMaxChangeRate, &v.MinSelfDelegation, &v.UpdatedAt, &v.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "validator not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

// GET /validators/:address/history
func GetValidatorHistory(c *gin.Context) {
	address := c.Param("address")
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM validator_history WHERE operator_address = $1`, address).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT id, operator_address, height, status, jailed, power, tokens,
			changed_by, change_type, old_power, new_power, timestamp, created_at
		FROM validator_history WHERE operator_address = $1
		ORDER BY height DESC LIMIT $2 OFFSET $3`, address, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	history := []models.ValidatorHistory{}
	for rows.Next() {
		var h models.ValidatorHistory
		rows.Scan(&h.ID, &h.OperatorAddress, &h.Height, &h.Status, &h.Jailed, &h.Power,
			&h.Tokens, &h.ChangedBy, &h.ChangeType, &h.OldPower, &h.NewPower, &h.Timestamp, &h.CreatedAt)
		history = append(history, h)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: history, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /validators/:address/stats
func GetValidatorStats(c *gin.Context) {
	address := c.Param("address")
	var s models.ValidatorStats
	err := config.DB.QueryRow(`
		SELECT operator_address, moniker, status, jailed, power, tokens,
			added_by, added_at, total_changes, last_change, blocks_proposed, updated_at
		FROM validator_stats WHERE operator_address = $1`, address).
		Scan(&s.OperatorAddress, &s.Moniker, &s.Status, &s.Jailed, &s.Power,
			&s.Tokens, &s.AddedBy, &s.AddedAt, &s.TotalChanges, &s.LastChange, &s.BlocksProposed, &s.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "validator stats not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GET /validator-stats
func ListValidatorStats(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM validator_stats`).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT operator_address, moniker, status, jailed, power, tokens,
			added_by, added_at, total_changes, last_change, blocks_proposed, updated_at
		FROM validator_stats ORDER BY power DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	stats := []models.ValidatorStats{}
	for rows.Next() {
		var s models.ValidatorStats
		rows.Scan(&s.OperatorAddress, &s.Moniker, &s.Status, &s.Jailed, &s.Power,
			&s.Tokens, &s.AddedBy, &s.AddedAt, &s.TotalChanges, &s.LastChange, &s.BlocksProposed, &s.UpdatedAt)
		stats = append(stats, s)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: stats, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}
