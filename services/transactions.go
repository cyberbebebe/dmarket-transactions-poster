package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
type Transaction struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	CustomID  string `json:"customId"`
	Emitter   string `json:"emitter"`
	Action    string `json:"action"`
	Subject   string `json:"subject"`
	Contractor struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"contractor"`
	Details struct {
    SettlementTime int64  `json:"-"`
    Image          string `json:"-"`
    ItemID         string `json:"itemId"`
    Extra          struct {
        FloatPartValue string  `json:"floatPartValue"`
        FloatValue     float64 `json:"floatValue"`
        PaintSeed      int     `json:"paintSeed"`
    	} `json:"extra"`
	} `json:"details"`
	Changes []struct {
		Money struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"money"`
		ChangeType string `json:"changeType"`
	} `json:"changes"`
	From      string `json:"from"`
	To        string `json:"to"`
	Status    string `json:"status"`
	Balance   struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"balance"`
	UpdatedAt int64 `json:"updatedAt"`
}


type TransactionsResponse struct {
	Objects []Transaction `json:"objects"`
	Total   int           `json:"total"`
}

var bot *tgbotapi.BotAPI

func GetLastTransactions(keysStamps map[string]string) (map[string][]Transaction){
	lastTransactions := make(map[string][]Transaction)
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
			fmt.Println("req err:", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil{
			fmt.Println("read body err:", err)
			continue
		}

		var response TransactionsResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Println("unmarshal err:", err)
			continue
		}

		if len(response.Objects) > 0 {
			fmt.Println("New trans for", key[64:])
			lastTransactions[key] = response.Objects
			// update last timestamp for this secretKey
			maxUpdatedAt := response.Objects[0].UpdatedAt
			keysStamps[key] = fmt.Sprintf("%d", maxUpdatedAt+1)
		}
		
	}
	return lastTransactions	
}

func PostTransactions(transactions map[string][]Transaction, chatIDs map[string]string) {
	
	for key, txs := range transactions {
		chatID, exists := chatIDs[key[64:]]
		if !exists {
			fmt.Println("No chat ID for key", key[64:])
			continue
		}

		for _, tx := range txs {
			moneySign := "-"
			statusFix := fixMarkdownV2(tx.Status)
			balance, _ := strconv.ParseFloat(tx.Balance.Amount, 64)
			change, _ := strconv.ParseFloat(tx.Changes[0].Money.Amount, 64)
			
			if tx.Action == "Sell" {
				moneySign = "+"
				deduction := math.Round(change * 0.02 * 100) / 100
				balance = balance - deduction
			}
			
			changeRounded := fmt.Sprintf("%.2f", change)
			balanceRounded := fmt.Sprintf("%.2f", balance)

			var floatLine string
			if tx.Details.Extra.FloatValue != 0{
				floatLine = fmt.Sprintf("Float: %.8f\n\n", tx.Details.Extra.FloatValue)
    		}
			message := fmt.Sprintf("%s %s\n`%s`\n\n%sMoney: %s%s $\nBalance: %s $",
				tx.Action,
				statusFix,
				tx.Subject,
				floatLine,
				moneySign,
				changeRounded,
				balanceRounded,
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
    replacements := map[string]string{
        "_": "\\_",
        "*": "\\*",
        "[": "\\[",
        "]": "\\]",
        "(": "\\(",
        ")": "\\)",
        "~": "\\~",
        "`": "\\`",
        ">": "\\>",
        "#": "\\#",
        "+": "\\+",
        "-": "\\-",
        "=": "\\=",
        "|": "\\|",
        "{": "\\{",
        "}": "\\}",
        ".": "\\.",
        "!": "\\!",
    }
    for from, to := range replacements {
        text = strings.ReplaceAll(text, from, to)
    }
    return text
}