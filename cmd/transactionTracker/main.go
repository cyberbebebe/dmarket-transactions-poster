package main

import (
	"fmt"
	"time"
	"github.com/cyberbebebe/dmarket-transactions-poster/services"
)

func main() {
	secretKeys, _ := services.GetSecretKeys() // Return data in [secretKey1, secretKey2, secretKey3...]
	keysStamps := services.MakeKeysStamps(secretKeys) // Return data in map[secretKey]string
	chatIDs, _ := services.LoadTransactionChatIDs() // Returns data in map[secretKey[64:]]string
	if err := services.InitTelegramBot(); err != nil {
		fmt.Println("Error initializing Telegram bot:", err)
		return
	}
	fmt.Println("Starting main cycle")
	for {
		lastTransactions := services.GetLastTransactions(keysStamps)
		services.PostTransactions(lastTransactions, chatIDs)
		time.Sleep(5 * time.Second)
	}
}

