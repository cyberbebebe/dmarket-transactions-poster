package services

import (
	"fmt"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WakeUpBots initializes a bot instance for every unique token in the config.
func WakeUpBots(configs []types.AccountConfig) (map[string]*tgbotapi.BotAPI, error) {
	botMap := make(map[string]*tgbotapi.BotAPI)

	fmt.Println("Waking up Telegram Bots...")

	for _, cfg := range configs {
		token := cfg.TelegramToken
		if token == "" {
			continue
		}

		// Only wake up if we haven't already
		if _, exists := botMap[token]; !exists {
			bot, err := tgbotapi.NewBotAPI(token)
			
			// --- ERROR HANDLING WITH PAUSE ---
			if err != nil {
				fmt.Printf("\nCRITICAL ERROR: Failed to init bot for account '%s'\n", cfg.Label)
				fmt.Printf("Reason: %v\n", err)
				fmt.Println("(Likely a wrong Telegram Token in config.json)")
				
				fmt.Println("\nPress [ENTER] to exit program...")
				var input string
				fmt.Scanln(&input) // <--- The Pause You Requested
				
				return nil, err
			}
			// ---------------------------------

			botMap[token] = bot
			fmt.Printf("   > Bot live: %s\n", bot.Self.UserName)
		}
	}
	return botMap, nil
}