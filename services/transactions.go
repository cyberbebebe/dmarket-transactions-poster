package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

func GetAllTransactions(secretKeys []string) map[string]map[string]float64 {
	allTransactions := make(map[string]map[string]float64)
	method := "GET"
	rootApiUrl := "https://api.dmarket.com"
	client := &http.Client{}

	fmt.Println("Loading purchase(s) history...")

	for _, key := range secretKeys {
		// Initialize inner map
		if _, exists := allTransactions[key]; !exists {
			allTransactions[key] = make(map[string]float64)
		}
		
		// totalLimit := 100000 // Limit how much buy transactions you want to check

		cursor := "" // Start with empty cursor
		keepFetching := true
		
		for keepFetching {
			// Construct URL with Cursor
			endpoint := fmt.Sprintf("/marketplace-api/v1/user-targets/closed?Limit=500&Status=successful,trade_protected&Cursor=%s", cursor)
			
			headers, _ := generateHeaders(key, method, endpoint, nil)
			req, _ := http.NewRequest(method, rootApiUrl+endpoint, nil)
			req.Header = headers

			resp, err := client.Do(req)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}

			// Handle Rate Limits
			if resp.StatusCode == 429 {
				resp.Body.Close()
				time.Sleep(5 * time.Second)
				continue 
			}
			
			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				fmt.Printf("API Error %d: %s\n", resp.StatusCode, string(body))
				break
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
					allTransactions[key][trade.AssetID] = trade.Price.Amount
				}
			}

			// Pagination
			if response.Cursor == "" {
				keepFetching = false
			} else {
				// Update cursor and little time sleep
				cursor = response.Cursor
				time.Sleep(100 * time.Millisecond)
			}
		}
		fmt.Printf("Total loaded buy transaction for key ...%s: %d items\n", key[64:96], len(allTransactions[key]))
	}

	return allTransactions
}

func GetLastTransactions(keysStamps map[string]string) (map[string][]types.Transaction){
	lastTransactions := make(map[string][]types.Transaction)
	method := "GET"
	rootApiUrl := "https://api.dmarket.com"
	client := &http.Client{}

	for key, timestamp := range keysStamps{
		endpoint := fmt.Sprintf("/exchange/v1/history?version=V3&from=%s&to=0&activities=sell,purchase,target_closed&statuses=success,trade_protected,reverted&offset=0&limit=10", timestamp) 
		headers, _ := generateHeaders(key, method, endpoint, nil)
		
		req, _ := http.NewRequest(method, rootApiUrl+endpoint, nil)
		
		req.Header = headers

		resp, err := client.Do(req)
		if err != nil{
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil{
			fmt.Println("read body err:", err)
			continue
		}

		var response types.TransactionsResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Println("unmarshal err:", err)
			continue
		}

		if len(response.Objects) > 0 {
			fmt.Println("New trans for", key[64:96])
			lastTransactions[key] = response.Objects
			// update last timestamp for this secretKey
			maxUpdatedAt := response.Objects[0].UpdatedAt
			keysStamps[key] = fmt.Sprintf("%d", maxUpdatedAt+1)
		}
		
	}
	return lastTransactions	
}

func PostTransactions(transactions map[string][]types.Transaction, allTransactions map[string]map[string]float64, chatIDs map[string]string) {
	
	for key, txs := range transactions {
		chatID, exists := chatIDs[key[64:]]
		if !exists {
			fmt.Println("No chat ID for key", key[64:])
			continue
		}

		if _, ok := allTransactions[key]; !ok {
			allTransactions[key] = make(map[string]float64)
		}

		for _, tx := range txs {
			var metaData strings.Builder
			var moneyData strings.Builder

			moneySign := "-"
			statusFix := fixMarkdownV2(tx.Status)
			balance, _ := strconv.ParseFloat(tx.Balance.Amount, 64)
			change, _ := strconv.ParseFloat(tx.Changes[0].Money.Amount, 64)

			profit := 0.0
			profitSign := ""
			showProfit := false
			
			// save bought items to dict
			if tx.Type == "purchase" || tx.Type == "target_closed" {
				if tx.Details.ItemID != "" {
					allTransactions[key][tx.Details.ItemID] = change
				}
			}

			if tx.Action == "Sell" {
				moneySign = "+"
				deduction := math.Round((change * 0.02) * 100) / 100 // fee
				balance = balance - deduction

				if tx.Status == "trade_protected"{
					balance = balance - change // should calculate usable balance
				}

				itemID := tx.Details.ItemID

				if itemID != "" {
					if buyPrice, found := allTransactions[key][itemID]; found {
						profit = change - buyPrice
						showProfit = true
						
						profitSign = "-"
						if profit >= 0 {
							profitSign = "+"
						}
					}
				}
			}

			changeRounded := fmt.Sprintf("%.2f", change)
			balanceRounded := fmt.Sprintf("%.2f", balance)
			
			moneyData.WriteString(fmt.Sprintf("Change: %s %s $", moneySign, changeRounded))
			if showProfit{
				moneyData.WriteString(fmt.Sprintf("\nProfit: %s %.2f $", profitSign, math.Abs(profit)))
			}
			moneyData.WriteString(fmt.Sprintf("\nBalance: %s $", balanceRounded))

			if tx.Details.Extra.FloatValue != 0.0{
				metaData.WriteString(fmt.Sprintf("\n\nFloat: %.8f", tx.Details.Extra.FloatValue))
    		}
			if tx.Details.Extra.PhaseTitle != ""{
				metaData.WriteString(fmt.Sprintf("\nPhase: %s", tx.Details.Extra.PhaseTitle))
			} 
			if tx.Details.Extra.PaintSeed != nil {
				metaData.WriteString(fmt.Sprintf("\nPattern: %d", *tx.Details.Extra.PaintSeed))
			}
			
			detailsBlock := metaData.String()
			moneyBlock := moneyData.String()

			message := fmt.Sprintf("%s %s\n`%s`%s\n\n%s",
				tx.Action,
				statusFix,
				tx.Subject,
				detailsBlock,
				moneyBlock,
			)

			if err := PostToTelegram(chatID, message); err != nil {
				fmt.Printf("Error posting to Telegram for key %s: %v\n", key, err)
			}
		}
	}
}

func PostToTelegram(chatID, message string) error {
	if bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}
	msg := tgbotapi.NewMessageToChannel(chatID, message)
	msg.ParseMode = "Markdown" // Enable Markdown for backticks
	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}
	return nil
}

func InitTelegramBot() error {
	token, err := GetTelegramBotToken()
	if err != nil {
		return fmt.Errorf("failed to get Telegram bot token: %w", err)
	}
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("failed to initialize Telegram bot: %w", err)
	}
	return nil
}

func fixMarkdownV2(text string) string {
    replacer := strings.NewReplacer(
        "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)", 
        "~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", 
        "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
    )
    return replacer.Replace(text)
}