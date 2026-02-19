package services

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartTracker is the main loop for a single DMarket account.
func StartTracker(cfg types.AccountConfig, bot *tgbotapi.BotAPI, costs types.CostMap, mu *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("[%s] Tracker Started\n", cfg.Label)

	lastTime := time.Now().Unix()

	for {
		// 1. Fetch History
		newTxs, nextTime, err := FetchNewTransactions(cfg.DMarketKey, lastTime)
		if err != nil {
			time.Sleep(15 * time.Second)
			continue
		}

		if len(newTxs) > 0 {
			// 2. Fetch Balance (Only if we have new txs)
			var currentBalance types.UserBalanceResponse
			
			if cfg.AdvancedBalance {
				currentBalance, _ = FetchUserBalance(cfg.DMarketKey)
				}

			for _, tx := range newTxs {

				if cfg.IgnoreReleased {

					// Skip success transactions that were trade protected, if true in config
					isOldTrade := tx.UpdatedAt > tx.CreatedAt 

					if tx.Status == "success" && isOldTrade {
						continue
					}
				}

				// Update Cost Map if we bought something
				if tx.Type == "target_closed" || tx.Type == "purchase" {
					amount, _ := strconv.ParseFloat(tx.Changes[0].Money.Amount, 64)
					if tx.Details.ItemID != "" {
						mu.Lock()
						costs[tx.Details.ItemID] = amount
						mu.Unlock()
					}
				}

				// Post it
				PostTransaction(bot, tx, cfg, costs, mu, currentBalance)
			}
			lastTime = nextTime
		}

		time.Sleep(15 * time.Second)
	}
}

// PostTransaction handles formatting and sending
func PostTransaction(bot *tgbotapi.BotAPI, tx types.Transaction, cfg types.AccountConfig, costs types.CostMap, mu *sync.RWMutex, liveBalance types.UserBalanceResponse) {
	
	// 1. Prepare Builders
	var metaData strings.Builder
	var moneyData strings.Builder

	// 2. Parse Basic Data
	change, _ := strconv.ParseFloat(tx.Changes[0].Money.Amount, 64)
	
	// Balance Logic (Snapshot vs Live)
	var balanceVal float64
	var pendingVal float64

	if cfg.AdvancedBalance && liveBalance.Usd != "" {
		b, _ := strconv.ParseFloat(liveBalance.Usd, 64)
		p, _ := strconv.ParseFloat(liveBalance.UsdTradeProtected, 64)
		balanceVal = b / 100
		pendingVal = p / 100
	} else {
		b, _ := strconv.ParseFloat(tx.Balance.Amount, 64)
		balanceVal = b
	}

	// 3. Logic: Signs, Fees, and Profit
	moneySign := "-"
	statusFix := fixMarkdownV2(tx.Status)
	showProfit := false
	profit := 0.0
	profitP := 0.0
	profitSign := ""

	if tx.Action == "Sell" {
		moneySign = "+"
		
		// Fee Deduction (Only needed if NOT using advanced/live balance)
		if !cfg.AdvancedBalance {
			deduction := math.Round((change * 0.02) * 100) / 100
			balanceVal = balanceVal - deduction
			if tx.Status == "trade_protected" {
				balanceVal = balanceVal - change
			}
		}

		// Calculate Profit
		if tx.Details.ItemID != "" {
			mu.RLock()
			buyPrice, found := costs[tx.Details.ItemID]
			mu.RUnlock()

			if found && buyPrice > 0 {
				profit = change - buyPrice
				profitP = (profit / buyPrice) * 100
				showProfit = true
				
				profitSign = "-"
				if profit >= 0 { profitSign = "+" }
			}
		}
	}

	// 4. Build "Details Block" (The middle part)
	if tx.Details.Extra.FloatValue != 0.0 {
		metaData.WriteString(fmt.Sprintf("\n\nFloat: %.8f", tx.Details.Extra.FloatValue))
	}
	if tx.Details.Extra.PhaseTitle != "" {
		metaData.WriteString(fmt.Sprintf("\nPhase: %s", tx.Details.Extra.PhaseTitle))
	}
	if tx.Details.Extra.PaintSeed != nil {
		metaData.WriteString(fmt.Sprintf("\nPattern: %d", *tx.Details.Extra.PaintSeed))
	}

	// 5. "Money Block"
	// Change: + 25.00 $
	moneyData.WriteString(fmt.Sprintf("Change: %s %.2f $", moneySign, change))

	// Profit: + 5.00 $ (+ 20.0 %)
	if showProfit {
		profitStr := fmt.Sprintf("\nProfit: %s %.2f $", profitSign, math.Abs(profit))
		if cfg.ProfitPercent {
			profitStr += fmt.Sprintf(" / %s %.2f %%", profitSign, math.Abs(profitP))
		}
		moneyData.WriteString(profitStr)
	}

	// Balance: 100.00 $ / 50.00 $
	balanceStr := fmt.Sprintf("\nBalance: %.2f $", balanceVal)
	if cfg.AdvancedBalance && pendingVal > 0 {
		balanceStr += fmt.Sprintf(" / %.2f $", pendingVal)
	}
	moneyData.WriteString(balanceStr)

	detailsBlock := metaData.String()
	moneyBlock := moneyData.String()

	// 6. Final Assembly
	message := fmt.Sprintf("%s %s\n`%s`%s\n\n%s",
		tx.Action,
		statusFix,
		tx.Subject,
		detailsBlock,
		moneyBlock,
	)

	// 7. Send
	msg := tgbotapi.NewMessageToChannel(cfg.TelegramChatID, message)
	msg.ParseMode = "Markdown"
	go func() {
        if _, err := bot.Send(msg); err != nil {
            fmt.Printf("[%s] Telegram Error: %v\n", cfg.Label, err)
        }
    }()
}


func fixMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)", 
		"~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", 
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
	)
	return replacer.Replace(text)
}