package handlers

import (
	"blockchain-api/config"
	"blockchain-api/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /proposals
func ListProposals(c *gin.Context) {
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status := c.Query("status")

	countQ := `SELECT COUNT(*) FROM proposals`
	listQ := `SELECT proposal_id, title, description, proposal_type, status, submit_time, deposit_end_time,
		voting_start_time, voting_end_time, total_deposit, deposit_denom, metadata, messages,
		yes_votes, no_votes, abstain_votes, no_with_veto_votes, proposer, height, updated_at, created_at
		FROM proposals`

	var total int64
	if status != "" {
		config.DB.QueryRow(countQ+` WHERE status = $1`, status).Scan(&total)
		rows, err := config.DB.Query(listQ+` WHERE status = $1 ORDER BY proposal_id DESC LIMIT $2 OFFSET $3`, status, pq.Limit(), pq.Offset())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		proposals := scanProposals(rows)
		c.JSON(http.StatusOK, models.PaginatedResponse{Data: proposals, Page: pq.Page, PerPage: pq.Limit(), Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit())))})
		return
	}

	config.DB.QueryRow(countQ).Scan(&total)
	rows, err := config.DB.Query(listQ+` ORDER BY proposal_id DESC LIMIT $1 OFFSET $2`, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	proposals := scanProposals(rows)
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: proposals, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /proposals/:id
func GetProposal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proposal id"})
		return
	}

	var p models.Proposal
	err = config.DB.QueryRow(`SELECT proposal_id, title, description, proposal_type, status, submit_time, deposit_end_time,
		voting_start_time, voting_end_time, total_deposit, deposit_denom, metadata, messages,
		yes_votes, no_votes, abstain_votes, no_with_veto_votes, proposer, height, updated_at, created_at
		FROM proposals WHERE proposal_id = $1`, id).
		Scan(&p.ProposalID, &p.Title, &p.Description, &p.ProposalType, &p.Status,
			&p.SubmitTime, &p.DepositEndTime, &p.VotingStartTime, &p.VotingEndTime,
			&p.TotalDeposit, &p.DepositDenom, &p.Metadata, &p.Messages,
			&p.YesVotes, &p.NoVotes, &p.AbstainVotes, &p.NoWithVetoVotes,
			&p.Proposer, &p.Height, &p.UpdatedAt, &p.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proposal not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GET /proposals/:id/votes
func GetProposalVotes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proposal id"})
		return
	}
	var pq models.PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total int64
	config.DB.QueryRow(`SELECT COUNT(*) FROM votes WHERE proposal_id = $1`, id).Scan(&total)

	rows, err := config.DB.Query(`
		SELECT id, proposal_id, voter, option, options, height, tx_hash, timestamp, created_at
		FROM votes WHERE proposal_id = $1 ORDER BY height DESC LIMIT $2 OFFSET $3`,
		id, pq.Limit(), pq.Offset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	votes := []models.Vote{}
	for rows.Next() {
		var v models.Vote
		rows.Scan(&v.ID, &v.ProposalID, &v.Voter, &v.Option, &v.Options, &v.Height, &v.TxHash, &v.Timestamp, &v.CreatedAt)
		votes = append(votes, v)
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: votes, Page: pq.Page, PerPage: pq.Limit(),
		Total: total, TotalPages: int(math.Ceil(float64(total) / float64(pq.Limit()))),
	})
}

// GET /addresses/:address/votes
func GetAddressVotes(c *gin.Context) {
	address := c.Param("address")
	rows, err := config.DB.Query(`
		SELECT id, proposal_id, voter, option, options, height, tx_hash, timestamp, created_at
		FROM votes WHERE voter = $1 ORDER BY height DESC`, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	votes := []models.Vote{}
	for rows.Next() {
		var v models.Vote
		rows.Scan(&v.ID, &v.ProposalID, &v.Voter, &v.Option, &v.Options, &v.Height, &v.TxHash, &v.Timestamp, &v.CreatedAt)
		votes = append(votes, v)
	}
	c.JSON(http.StatusOK, gin.H{"address": address, "votes": votes})
}

func scanProposals(rows interface {
	Next() bool
	Scan(...interface{}) error
}) []models.Proposal {
	proposals := []models.Proposal{}
	for rows.Next() {
		var p models.Proposal
		rows.Scan(&p.ProposalID, &p.Title, &p.Description, &p.ProposalType, &p.Status,
			&p.SubmitTime, &p.DepositEndTime, &p.VotingStartTime, &p.VotingEndTime,
			&p.TotalDeposit, &p.DepositDenom, &p.Metadata, &p.Messages,
			&p.YesVotes, &p.NoVotes, &p.AbstainVotes, &p.NoWithVetoVotes,
			&p.Proposer, &p.Height, &p.UpdatedAt, &p.CreatedAt)
		proposals = append(proposals, p)
	}
	return proposals
}
