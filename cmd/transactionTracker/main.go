package main

import (
	"fmt"
	"sync"

	"github.com/cyberbebebe/dmarket-transactions-poster/services"
)

func main() {
	// 1. Load Config
	configs, err := services.LoadConfig("config/config.json")
	if err != nil {
		panic(err)
	}

	// 2. Prepare Data (The Brain)
	costMap, costMu := services.InitCostBasis(configs)

	// 3. Wake up telegram bots
	botMap, err := services.WakeUpBots(configs)
	if err != nil { panic(err) }
	
	// 4. Start Workers
	var wg sync.WaitGroup

	fmt.Println("Launching Workers...")

	for _, cfg := range configs {
		wg.Add(1)
		// Launch a Tracker for each account
		botInstance := botMap[cfg.TelegramToken]
		
		go services.StartTracker(cfg, botInstance, costMap, costMu, &wg)

		if cfg.CSFloatKey != "" {
			wg.Add(1)
			go services.StartCSFloatPoller(cfg, costMap, costMu, &wg)
		}
	}

	wg.Wait()
}