package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
)

// InitCostBasis loads ALL buy history (DMarket + CSFloat) for all accounts
func InitCostBasis(configs []types.AccountConfig) (types.CostMap, *sync.RWMutex) {
	
	// 1. Initialize the Shared Brain
	costMap := make(types.CostMap)
	var mu sync.RWMutex

	fmt.Println("Initializing Cost Basis...")

	// 2. Load DMarket History (Direct Buys)
	for _, cfg := range configs {
		dmCosts, err := FetchDMarketBuyHistory(cfg.DMarketKey) 
		
		if err != nil {
			fmt.Printf("⚠️ Error fetching DMarket history for %s: %v\n", cfg.Label, err)
		} else {
			// Write to Shared Map safely
			mu.Lock()
			for id, price := range dmCosts {
				costMap[id] = price
			}
			mu.Unlock()
			fmt.Printf("Loaded %d DMarket buys\n", len(dmCosts))
		}
		
		// 3. Sync CSFloat (If key exists)
		if cfg.CSFloatKey != "" {
			SyncCSFloatCosts(cfg, costMap, &mu)
		}
	}

	fmt.Printf("Total Tracked Items: %d\n", len(costMap))
	return costMap, &mu
}

func FetchDMarketBuyHistory(secretKey string) (map[string]float64, error) {
	transactions := make(map[string]float64)
	method := "GET"
	rootApiUrl := "https://api.dmarket.com"
	client := &http.Client{}

	fmt.Println("Loading purchase(s) history...")

	cursor := "" // Start with empty cursor
	keepFetching := true
		
	for keepFetching {
		// URL
		endpoint := fmt.Sprintf("/marketplace-api/v1/user-targets/closed?Limit=500&OrderDir=asc&Status=successful,trade_protected&Cursor=%s", cursor)
			
		headers, _ := generateHeaders(secretKey, method, endpoint, nil)
		req, _ := http.NewRequest(method, rootApiUrl+endpoint, nil)
		req.Header = headers

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(5 * time.Second)
			continue 
		}
			
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("API Error %d: %s\n", resp.StatusCode, string(body))
		}
			
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		var response types.UserTargetsClosedResponse
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			break
		}

		// Process items
		for _, trade := range response.Trades {
			if trade.AssetID != "" {
				// Direct assignment to the map
				transactions[trade.AssetID] = trade.Price.Amount
			}
		}

		// Pagination
		if response.Cursor == "" {
				keepFetching = false
		} else {
			// Update cursor and time sleep
			cursor = response.Cursor
			time.Sleep(100 * time.Millisecond)
		}
		}
	
	return transactions, nil
}

func FetchCSFloatHistory(apiKey string) (map[string]float64, error) {
	buyHistory := make(map[string]float64)
	client := &http.Client{}
	
	// 'verified' and 'pending' trades
	baseUrl := "https://csfloat.com/api/v1/me/trades?role=buyer&state=verified,pending&limit=1000"
	page := 0
	
	fmt.Printf("Fetching CSFloat history (Page 0)...")

	for {
		// 1. Construct URL with pagination
		url := fmt.Sprintf("%s&page=%d", baseUrl, page)
		
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("network error: %v", err)
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			fmt.Println("\n⚠️ CSFloat Rate Limit. Sleeping 5s...")
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		var response types.CSFloatResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("unmarshal error: %v", err)
		}

		// 2. Stop condition: No more trades returned
		if len(response.Trades) == 0 {
			break
		}

		// 3. Process Trades
		for _, trade := range response.Trades {
			item := trade.Contract.Item
			
			// We need both Float and Seed to create a unique fingerprint
			if item.FloatValue > 0 {
				// Create Fingerprint: "0.12345678-555"
				// Must match the format used in Matcher function
				fingerprint := fmt.Sprintf("%f-%d", item.FloatValue, *item.PaintSeed)
				
				// Convert cents to dollars
				priceUSD := float64(trade.Contract.Price) / 100.0
				
				buyHistory[fingerprint] = priceUSD
			}
		}
		
		fmt.Printf(".") // Progress indicator
		page++
		time.Sleep(1 * time.Second) // Be polite to API
	}

	fmt.Println() // New line after dots
	return buyHistory, nil
}

func FetchDMarketInventory(secretKey string) ([]types.DMarketInventoryItem, error) {
	var inventory []types.DMarketInventoryItem
	
	method := "GET"
	rootApiUrl := "https://api.dmarket.com"
	client := &http.Client{}
	
	cursor := ""
	keepFetching := true
	
	// Privacy logging
	keyLog := "???"
	if len(secretKey) > 10 {
		keyLog = secretKey[len(secretKey)-10:]
	}
	fmt.Printf("Fetching DMarket inventory for key ...%s\n", keyLog)

	for keepFetching {
		// Use the correct endpoint for "user offers" (Inventory/On Sale)
		endpoint := fmt.Sprintf("/exchange/v1/user/offers?side=user&orderBy=price&orderDir=desc&gameId=a8db&limit=100&currency=USD&cursor=%s", cursor)
		
		headers, _ := generateHeaders(secretKey, method, endpoint, nil)
		req, _ := http.NewRequest(method, rootApiUrl+endpoint, nil)
		req.Header = headers

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		var response types.DMarketInventoryResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("unmarshal error: %v", err)
		}

		// Append items
		inventory = append(inventory, response.Objects...)

		// Pagination
		if response.Cursor == "" {
			keepFetching = false
		} else {
			cursor = response.Cursor
			// Slight delay
			time.Sleep(100 * time.Millisecond)
		}
	}

	return inventory, nil
}

// SyncCSFloatCosts fetches CSFloat history and matches it with DMarket inventory
func SyncCSFloatCosts(cfg types.AccountConfig, costs types.CostMap, mu *sync.RWMutex) {
	fmt.Printf("Syncing CSFloat Buys for %s...\n", cfg.Label)
	
	// 1. Fetch CSFloat Buy History (Map of "Float-Seed" -> Price)
	csfloatBuys, err := FetchCSFloatHistory(cfg.CSFloatKey)
	if err != nil {
		fmt.Printf("Error fetching CSFloat history: %v\n", err)
		return
	}

	// 2. Fetch Current DMarket Inventory (To get ItemIDs)
	inventory, err := FetchDMarketInventory(cfg.DMarketKey)
	if err != nil {
		fmt.Printf("Error fetching DMarket inventory: %v\n", err)
		return
	}

	// 3. MATCHING LOGIC
	matches := 0
	mu.Lock() 
	defer mu.Unlock()

	for _, item := range inventory {
		// Skip items with 0 float (stickers/cases/etc) unless you can track them
		if item.Extra.FloatValue == 0 {
			continue
		}

		seed := 0
        if item.Extra.PaintSeed != nil {
            seed = *item.Extra.PaintSeed
        }

		// Create the "Fingerprint", format: "0.011534607969224453-712"
		fingerprint := fmt.Sprintf("%f-%d", item.Extra.FloatValue, seed)

		// Check if we have a buy record for this fingerprint
		if price, found := csfloatBuys[fingerprint]; found {
			// We map the DMarket ItemID (from inventory) to the Price (from CSFloat)
			costs[item.ItemID] = price 
			matches++
		}
	}
	
	fmt.Printf("Matched %d CSFloat items to DMarket Inventory\n", matches)
}