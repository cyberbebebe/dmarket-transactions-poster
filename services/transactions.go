package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
)

// FetchNewTransactions gets history items strictly NEWER than lastTimestamp.
func FetchNewTransactions(secretKey string, lastTimestamp int64) ([]types.Transaction, int64, error) {
	var newTransactions []types.Transaction
	
	// We ask for the last 50 and filter manually
	endpoint := "/exchange/v1/history?version=V3&limit=50&activities=sell,purchase,target_closed&statuses=success,trade_protected,reverted"
	
	method := "GET"
	rootApiUrl := "https://api.dmarket.com"
	client := &http.Client{}

	headers, _ := generateHeaders(secretKey, method, endpoint, nil)
	req, _ := http.NewRequest(method, rootApiUrl+endpoint, nil)
	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return nil, lastTimestamp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, lastTimestamp, fmt.Errorf("API status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response types.TransactionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, lastTimestamp, err
	}

	// Filter for new items and update timestamp
	newestTS := lastTimestamp

	// DMarket returns newest first.
	for _, tx := range response.Objects {
		if tx.UpdatedAt > lastTimestamp {
			newTransactions = append(newTransactions, tx)
			if tx.UpdatedAt > newestTS {
				newestTS = tx.UpdatedAt
			}
		}
	}

	return newTransactions, newestTS, nil
}

// FetchUserBalance to get Real + Pending balance.
func FetchUserBalance(secretKey string) (types.UserBalanceResponse, error) {
	var balance types.UserBalanceResponse
	
	endpoint := "/account/v1/balance"
	rootApiUrl := "https://api.dmarket.com"
	client := &http.Client{}

	headers, _ := generateHeaders(secretKey, "GET", endpoint, nil)
	req, _ := http.NewRequest("GET", rootApiUrl+endpoint, nil)
	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return balance, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return balance, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &balance); err != nil {
		return balance, err
	}

	return balance, nil
}