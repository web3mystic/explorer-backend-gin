package models

import (
	"encoding/json"
	"time"
)

// NullableTime handles nullable timestamps
type NullableTime struct {
	Time  time.Time
	Valid bool
}

func (nt NullableTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type PaginationQuery struct {
	Page    int `form:"page,default=1"`
	PerPage int `form:"per_page,default=20"`
}

func (p PaginationQuery) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PerPage
}

func (p PaginationQuery) Limit() int {
	if p.PerPage < 1 || p.PerPage > 100 {
		return 20
	}
	return p.PerPage
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// ─── Block ────────────────────────────────────────────────────────────────────

type Block struct {
	Height          int64     `json:"height"`
	Hash            string    `json:"hash"`
	Time            time.Time `json:"time"`
	ProposerAddress string    `json:"proposer_address"`
	TxCount         int       `json:"tx_count"`
	CreatedAt       time.Time `json:"created_at"`
}

// ─── Transaction ──────────────────────────────────────────────────────────────

type Transaction struct {
	Hash      string          `json:"hash"`
	Height    int64           `json:"height"`
	TxIndex   int             `json:"tx_index"`
	GasUsed   int64           `json:"gas_used"`
	GasWanted int64           `json:"gas_wanted"`
	Fee       string          `json:"fee"`
	Success   bool            `json:"success"`
	Code      int             `json:"code"`
	Log       string          `json:"log"`
	Memo      string          `json:"memo"`
	MsgTypes  json.RawMessage `json:"msg_types"`
	Addresses json.RawMessage `json:"addresses"`
	Timestamp time.Time       `json:"timestamp"`
	CreatedAt time.Time       `json:"created_at"`
}

type RecentTransaction struct {
	Hash      string    `json:"hash"`
	Height    int64     `json:"height"`
	TxIndex   int       `json:"tx_index"`
	Success   bool      `json:"success"`
	GasUsed   int64     `json:"gas_used"`
	GasWanted int64     `json:"gas_wanted"`
	Fee       string    `json:"fee"`
	Memo      string    `json:"memo"`
	Timestamp time.Time `json:"timestamp"`
	BlockTime time.Time `json:"block_time"`
}

// ─── Address / Balance ────────────────────────────────────────────────────────

type AddressTransaction struct {
	Address   string    `json:"address"`
	TxHash    string    `json:"tx_hash"`
	Height    int64     `json:"height"`
	CreatedAt time.Time `json:"created_at"`
}

type Balance struct {
	Address   string    `json:"address"`
	Denom     string    `json:"denom"`
	Amount    string    `json:"amount"`
	Height    int64     `json:"height"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Bank Transfer ────────────────────────────────────────────────────────────

type BankTransfer struct {
	ID          int64     `json:"id"`
	TxHash      string    `json:"tx_hash"`
	MsgIndex    int       `json:"msg_index"`
	Height      int64     `json:"height"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      string    `json:"amount"`
	Denom       string    `json:"denom"`
	AmountValue float64   `json:"amount_value"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Validator ────────────────────────────────────────────────────────────────

type Validator struct {
	OperatorAddress       string     `json:"operator_address"`
	ConsensusAddress      string     `json:"consensus_address"`
	ConsensusPubkey       string     `json:"consensus_pubkey"`
	Moniker               string     `json:"moniker"`
	Identity              string     `json:"identity"`
	Website               string     `json:"website"`
	SecurityContact       string     `json:"security_contact"`
	Details               string     `json:"details"`
	Jailed                bool       `json:"jailed"`
	Status                string     `json:"status"`
	Power                 int64      `json:"power"`
	VotingPower           int64      `json:"voting_power"`
	Tokens                string     `json:"tokens"`
	DelegatorShares       string     `json:"delegator_shares"`
	AddedBy               string     `json:"added_by"`
	AddedAt               *time.Time `json:"added_at"`
	AddedAtHeight         int64      `json:"added_at_height"`
	UnbondingHeight       int64      `json:"unbonding_height"`
	UnbondingTime         *time.Time `json:"unbonding_time"`
	CommissionRate        string     `json:"commission_rate"`
	CommissionMaxRate     string     `json:"commission_max_rate"`
	CommissionMaxChangeRate string   `json:"commission_max_change_rate"`
	MinSelfDelegation     string     `json:"min_self_delegation"`
	UpdatedAt             time.Time  `json:"updated_at"`
	CreatedAt             time.Time  `json:"created_at"`
}

type ValidatorStats struct {
	OperatorAddress string     `json:"operator_address"`
	Moniker         string     `json:"moniker"`
	Status          string     `json:"status"`
	Jailed          bool       `json:"jailed"`
	Power           int64      `json:"power"`
	Tokens          string     `json:"tokens"`
	AddedBy         string     `json:"added_by"`
	AddedAt         *time.Time `json:"added_at"`
	TotalChanges    int64      `json:"total_changes"`
	LastChange      *time.Time `json:"last_change"`
	BlocksProposed  int64      `json:"blocks_proposed"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ValidatorHistory struct {
	ID              int        `json:"id"`
	OperatorAddress string     `json:"operator_address"`
	Height          int64      `json:"height"`
	Status          string     `json:"status"`
	Jailed          bool       `json:"jailed"`
	Power           int64      `json:"power"`
	Tokens          string     `json:"tokens"`
	ChangedBy       string     `json:"changed_by"`
	ChangeType      string     `json:"change_type"`
	OldPower        int64      `json:"old_power"`
	NewPower        int64      `json:"new_power"`
	Timestamp       *time.Time `json:"timestamp"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ─── Proposal / Governance ────────────────────────────────────────────────────

type Proposal struct {
	ProposalID      int64           `json:"proposal_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	ProposalType    string          `json:"proposal_type"`
	Status          string          `json:"status"`
	SubmitTime      *time.Time      `json:"submit_time"`
	DepositEndTime  *time.Time      `json:"deposit_end_time"`
	VotingStartTime *time.Time      `json:"voting_start_time"`
	VotingEndTime   *time.Time      `json:"voting_end_time"`
	TotalDeposit    string          `json:"total_deposit"`
	DepositDenom    string          `json:"deposit_denom"`
	Metadata        string          `json:"metadata"`
	Messages        json.RawMessage `json:"messages"`
	YesVotes        string          `json:"yes_votes"`
	NoVotes         string          `json:"no_votes"`
	AbstainVotes    string          `json:"abstain_votes"`
	NoWithVetoVotes string          `json:"no_with_veto_votes"`
	Proposer        string          `json:"proposer"`
	Height          int64           `json:"height"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

type Vote struct {
	ID         int             `json:"id"`
	ProposalID int64           `json:"proposal_id"`
	Voter      string          `json:"voter"`
	Option     string          `json:"option"`
	Options    json.RawMessage `json:"options"`
	Height     int64           `json:"height"`
	TxHash     string          `json:"tx_hash"`
	Timestamp  *time.Time      `json:"timestamp"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ─── Events / Messages ────────────────────────────────────────────────────────

type Event struct {
	ID         int             `json:"id"`
	TxHash     string          `json:"tx_hash"`
	EventIndex int             `json:"event_index"`
	EventType  string          `json:"event_type"`
	Attributes json.RawMessage `json:"attributes"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Message struct {
	ID        int             `json:"id"`
	TxHash    string          `json:"tx_hash"`
	MsgIndex  int             `json:"msg_index"`
	MsgType   string          `json:"msg_type"`
	Sender    string          `json:"sender"`
	Receiver  string          `json:"receiver"`
	Amount    string          `json:"amount"`
	Denom     string          `json:"denom"`
	RawData   json.RawMessage `json:"raw_data"`
	CreatedAt time.Time       `json:"created_at"`
}

// ─── Authority Accounts ───────────────────────────────────────────────────────

type AuthorityAccount struct {
	Address       string     `json:"address"`
	AddedAt       *time.Time `json:"added_at"`
	AddedBy       string     `json:"added_by"`
	AddedAtHeight int64      `json:"added_at_height"`
	Active        bool       `json:"active"`
	RemovedAt     *time.Time `json:"removed_at"`
	RemovedBy     string     `json:"removed_by"`
	Source        string     `json:"source"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ─── Wasm ─────────────────────────────────────────────────────────────────────

type WasmCode struct {
	CodeID         int64      `json:"code_id"`
	Creator        string     `json:"creator"`
	Checksum       string     `json:"checksum"`
	Permission     string     `json:"permission"`
	UploadedHeight int64      `json:"uploaded_height"`
	UploadedTime   *time.Time `json:"uploaded_time"`
	UploadTxHash   string     `json:"upload_tx_hash"`
	CreatedAt      time.Time  `json:"created_at"`
}

type WasmContract struct {
	ContractAddress      string          `json:"contract_address"`
	CodeID               int64           `json:"code_id"`
	Creator              string          `json:"creator"`
	Admin                string          `json:"admin"`
	Label                string          `json:"label"`
	InitMsg              json.RawMessage `json:"init_msg"`
	ContractInfo         json.RawMessage `json:"contract_info"`
	InstantiatedAtHeight int64           `json:"instantiated_at_height"`
	InstantiatedAtTime   *time.Time      `json:"instantiated_at_time"`
	InstantiateTxHash    string          `json:"instantiate_tx_hash"`
	CurrentCodeID        int64           `json:"current_code_id"`
	LastMigratedHeight   int64           `json:"last_migrated_height"`
	LastMigratedTxHash   string          `json:"last_migrated_tx_hash"`
	IsActive             bool            `json:"is_active"`
	UpdatedAt            time.Time       `json:"updated_at"`
	CreatedAt            time.Time       `json:"created_at"`
}

type WasmExecution struct {
	ID              int64           `json:"id"`
	TxHash          string          `json:"tx_hash"`
	MsgIndex        int             `json:"msg_index"`
	Height          int64           `json:"height"`
	Sender          string          `json:"sender"`
	ContractAddress string          `json:"contract_address"`
	ExecuteMsg      json.RawMessage `json:"execute_msg"`
	ExecuteAction   string          `json:"execute_action"`
	Funds           json.RawMessage `json:"funds"`
	GasUsed         int64           `json:"gas_used"`
	Success         bool            `json:"success"`
	Error           string          `json:"error"`
	Timestamp       *time.Time      `json:"timestamp"`
	CreatedAt       time.Time       `json:"created_at"`
}

type WasmInstantiation struct {
	ID              int64           `json:"id"`
	TxHash          string          `json:"tx_hash"`
	MsgIndex        int             `json:"msg_index"`
	Height          int64           `json:"height"`
	Creator         string          `json:"creator"`
	Admin           string          `json:"admin"`
	CodeID          int64           `json:"code_id"`
	Label           string          `json:"label"`
	ContractAddress string          `json:"contract_address"`
	InitMsg         json.RawMessage `json:"init_msg"`
	Funds           json.RawMessage `json:"funds"`
	Success         bool            `json:"success"`
	Error           string          `json:"error"`
	Timestamp       *time.Time      `json:"timestamp"`
	CreatedAt       time.Time       `json:"created_at"`
}

type WasmMigration struct {
	ID              int64           `json:"id"`
	TxHash          string          `json:"tx_hash"`
	MsgIndex        int             `json:"msg_index"`
	Height          int64           `json:"height"`
	Sender          string          `json:"sender"`
	ContractAddress string          `json:"contract_address"`
	OldCodeID       int64           `json:"old_code_id"`
	NewCodeID       int64           `json:"new_code_id"`
	MigrateMsg      json.RawMessage `json:"migrate_msg"`
	Success         bool            `json:"success"`
	Error           string          `json:"error"`
	Timestamp       *time.Time      `json:"timestamp"`
	CreatedAt       time.Time       `json:"created_at"`
}

type WasmEvent struct {
	ID              int64           `json:"id"`
	TxHash          string          `json:"tx_hash"`
	MsgIndex        int             `json:"msg_index"`
	EventIndex      int             `json:"event_index"`
	Height          int64           `json:"height"`
	ContractAddress string          `json:"contract_address"`
	Action          string          `json:"action"`
	RawAttributes   json.RawMessage `json:"raw_attributes"`
	Timestamp       *time.Time      `json:"timestamp"`
	CreatedAt       time.Time       `json:"created_at"`
}

type ContractActivity struct {
	ContractAddress string     `json:"contract_address"`
	Label           string     `json:"label"`
	CodeID          int64      `json:"code_id"`
	Creator         string     `json:"creator"`
	TotalExecutions int64      `json:"total_executions"`
	LastExecution   *time.Time `json:"last_execution"`
	UniqueUsers     int64      `json:"unique_users"`
}

type RecentWasmActivity struct {
	TxHash          string     `json:"tx_hash"`
	Height          int64      `json:"height"`
	ContractAddress string     `json:"contract_address"`
	ContractLabel   string     `json:"contract_label"`
	Action          string     `json:"action"`
	Sender          string     `json:"sender"`
	Recipient       string     `json:"recipient"`
	Amount          string     `json:"amount"`
	Timestamp       *time.Time `json:"timestamp"`
}

// ─── CW20 ─────────────────────────────────────────────────────────────────────

type CW20Transfer struct {
	ID              int64           `json:"id"`
	TxHash          string          `json:"tx_hash"`
	MsgIndex        int             `json:"msg_index"`
	Height          int64           `json:"height"`
	ContractAddress string          `json:"contract_address"`
	Action          string          `json:"action"`
	FromAddress     string          `json:"from_address"`
	ToAddress       string          `json:"to_address"`
	Amount          float64         `json:"amount"`
	Memo            string          `json:"memo"`
	RawAttributes   json.RawMessage `json:"raw_attributes"`
	Timestamp       *time.Time      `json:"timestamp"`
	CreatedAt       time.Time       `json:"created_at"`
}

type CW20AddressActivity struct {
	ContractAddress string  `json:"contract_address"`
	Address         string  `json:"address"`
	TransfersSent   int64   `json:"transfers_sent"`
	TotalSent       float64 `json:"total_sent"`
}

// ─── Indexer / Sync State ─────────────────────────────────────────────────────

type IndexerState struct {
	ID            int       `json:"id"`
	LastHeight    int64     `json:"last_height"`
	LastBlockHash string    `json:"last_block_hash"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ValidatorSyncState struct {
	ID             int       `json:"id"`
	LastSyncHeight int64     `json:"last_sync_height"`
	LastSyncTime   time.Time `json:"last_sync_time"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ProposalSyncState struct {
	ID             int       `json:"id"`
	LastSyncHeight int64     `json:"last_sync_height"`
	LastSyncTime   time.Time `json:"last_sync_time"`
	UpdatedAt      time.Time `json:"updated_at"`
}
