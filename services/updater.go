package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
)

// StartCSFloatPoller periodically syncs CSFloat buys for a specific account.
func StartCSFloatPoller(cfg types.AccountConfig, costs types.CostMap, mu *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()
	
	fmt.Printf("[%s] CSFloat Auto-Updater Active\n", cfg.Label)

	for {
		// 1. Sleep for X minutes (e.g., 30 minutes)
		// We sleep FIRST because we already ran an initial sync in main.go/InitCostBasis
		time.Sleep(72 * time.Hour)

		// 2. Run Sync
		// This uses the function we wrote in cost_basis.go
		SyncCSFloatCosts(cfg, costs, mu)
	}
}