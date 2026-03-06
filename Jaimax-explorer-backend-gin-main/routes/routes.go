package routes

import (
	"blockchain-api/handlers"
	"blockchain-api/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
    ginSwagger   "github.com/swaggo/gin-swagger"
    _ "blockchain-api/docs"
)

func Register(r *gin.Engine) {
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api/v1")
	{
		// ── Health / Stats ──────────────────────────────────────────────────
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
		api.GET("/stats", handlers.GetStats)
		api.GET("/search", handlers.Search)

		// ── Blocks ──────────────────────────────────────────────────────────
		blocks := api.Group("/blocks")
		{
			blocks.GET("", handlers.ListBlocks)
			blocks.GET("/latest", handlers.GetLatestBlock)
			blocks.GET("/:height", handlers.GetBlock)
			blocks.GET("/:height/transactions", handlers.GetBlockTransactions)
			blocks.GET("/:height/transfers", handlers.GetBlockTransfers)
		}

		// ── Transactions ─────────────────────────────────────────────────────
		txs := api.Group("/transactions")
		{
			txs.GET("", handlers.ListTransactions)
			txs.GET("/recent", handlers.ListRecentTransactions)
			txs.GET("/:hash", handlers.GetTransaction)
			txs.GET("/:hash/messages", handlers.GetTransactionMessages)
			txs.GET("/:hash/events", handlers.GetTransactionEvents)
		}

		// ── Addresses ────────────────────────────────────────────────────────
		addr := api.Group("/addresses")
		{
			addr.GET("/:address", handlers.GetAddressSummary)              // ★ NEW
			addr.GET("/:address/balances", handlers.GetAddressBalances)
			addr.GET("/:address/transactions", handlers.GetAddressTransactions)
			addr.GET("/:address/bank-transfers", handlers.GetAddressBankTransfers)
			addr.GET("/:address/cw20-transfers", handlers.GetAddressCW20Transfers)
			addr.GET("/:address/votes", handlers.GetAddressVotes)
		}

		// ── Balances ─────────────────────────────────────────────────────────
		api.GET("/balances", handlers.ListBalances)
		api.GET("/balances/summary", handlers.GetBalancesSummary)          // ★ NEW

		// ── Bank Transfers ───────────────────────────────────────────────────
		bank := api.Group("/bank-transfers")
		{
			bank.GET("", handlers.ListBankTransfers)
			bank.GET("/:id", handlers.GetBankTransfer)
		}

		// ── Validators ───────────────────────────────────────────────────────
		vals := api.Group("/validators")
		{
			vals.GET("", handlers.ListValidators)
			vals.GET("/:address", handlers.GetValidator)
			vals.GET("/:address/history", handlers.GetValidatorHistory)
			vals.GET("/:address/stats", handlers.GetValidatorStats)
			vals.GET("/:address/blocks", handlers.GetValidatorBlocks)      // ★ NEW
		}
		api.GET("/validator-stats", handlers.ListValidatorStats)

		// ── Governance ───────────────────────────────────────────────────────
		gov := api.Group("/proposals")
		{
			gov.GET("", handlers.ListProposals)
			gov.GET("/:id", handlers.GetProposal)
			gov.GET("/:id/votes", handlers.GetProposalVotes)
		}

		// ── Wasm ─────────────────────────────────────────────────────────────
		wasm := api.Group("/wasm")
		{
			// Codes
			wasm.GET("/codes", handlers.ListWasmCodes)
			wasm.GET("/codes/:code_id", handlers.GetWasmCode)
			wasm.GET("/codes/:code_id/contracts", handlers.GetContractsByCode)
			wasm.GET("/codes/:code_id/instantiations", handlers.GetInstantiationsByCode) // ★ NEW

			// Contracts
			wasm.GET("/contracts", handlers.ListWasmContracts)
			wasm.GET("/contracts/:address", handlers.GetWasmContract)
			wasm.GET("/contracts/:address/executions", handlers.GetContractExecutions)
			wasm.GET("/contracts/:address/events", handlers.GetContractEvents)
			wasm.GET("/contracts/:address/migrations", handlers.GetContractMigrations)

			// Executions & Instantiations
			wasm.GET("/executions", handlers.ListWasmExecutions)
			wasm.GET("/executions/:id", handlers.GetWasmExecution)         // ★ NEW
			wasm.GET("/instantiations", handlers.ListWasmInstantiations)

			// Events
			wasm.GET("/events", handlers.ListWasmEvents)                   // ★ NEW

			// Activity
			wasm.GET("/activity", handlers.ListContractActivity)
			wasm.GET("/activity/recent", handlers.ListRecentWasmActivity)
		}

		// ── CW20 ─────────────────────────────────────────────────────────────
		cw20 := api.Group("/cw20")
		{
			cw20.GET("/transfers", handlers.ListCW20Transfers)
			cw20.GET("/address-activity", handlers.ListCW20AddressActivity)
		}

		// ── Authority Accounts ───────────────────────────────────────────────
		auth := api.Group("/authority-accounts")
		{
			auth.GET("", handlers.ListAuthorityAccounts)
			auth.GET("/:address", handlers.GetAuthorityAccount)
		}

		// ── Sync / Indexer State ─────────────────────────────────────────────
		state := api.Group("/state")
		{
			state.GET("/indexer", handlers.GetIndexerState)
			state.GET("/validator-sync", handlers.GetValidatorSyncState)
			state.GET("/proposal-sync", handlers.GetProposalSyncState)
		}
	}
}