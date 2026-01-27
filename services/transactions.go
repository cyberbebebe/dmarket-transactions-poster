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

	fmt.Println("Loading all historical transactions to build price cache...")

	for _, key := range secretKeys {
		// Initialize the inner map for this secret key
		if _, exists := allTransactions[key]; !exists {
			allTransactions[key] = make(map[string]float64)
		}

		offset := 0
		limit := 1000 // Batch size
		keepFetching := true

		for keepFetching {
			// Check "purchase" and "target_closed" to find BUY prices.
			endpoint := fmt.Sprintf("/exchange/v1/history?version=V3&from=0&activities=purchase,target_closed&statuses=success,trade_protected&offset=%d&limit=%d", offset, limit)
			
			headers, _ := generateHeaders(key, method, endpoint, nil)
			req, _ := http.NewRequest(method, rootApiUrl+endpoint, nil)
			req.Header = headers

			resp, err := client.Do(req)
			if err != nil {
				break 
			}
			
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fmt.Printf("Read body error: %v\n", err)
				break
			}

			var response types.TransactionsResponse
			if err := json.Unmarshal(body, &response); err != nil {
				fmt.Printf("Unmarshal error: %v\n", err)
				break
			}

			// Process this batch
			if len(response.Objects) == 0 {
				keepFetching = false
			} else {
				for _, tx := range response.Objects {
					// We only care if there is an ItemID and a valid money amount
					if tx.Details.ItemID != "" && len(tx.Changes) > 0 {
						amountStr := tx.Changes[0].Money.Amount
						price, err := strconv.ParseFloat(amountStr, 64)
						if err == nil {
							// Store the buy price for this ItemID
							allTransactions[key][tx.Details.ItemID] = price
						}
					}
				}
				
				// Set offset
				offset += len(response.Objects)
				
				// Slight delay
				time.Sleep(100 * time.Millisecond)
				
				// Stop if got less object than limit (means last page)
				if len(response.Objects) < limit {
					keepFetching = false
				}
			}
		}
		fmt.Printf("Loaded %d items for key %s\n", len(allTransactions[key]), key[64:])
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