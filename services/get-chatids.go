package services

import (
	"encoding/json"
	"fmt"
	"os"
)

// ChatIDConfig holds the structure of each entry in chatids.json
type ChatIDConfig struct {
	Offers      string `json:"offers"`
	Transactions string `json:"transactions"`
}

// LoadTransactionChatIDs loads chatids.json and returns a map of public keys to transaction chat IDs
func LoadTransactionChatIDs() (map[string]string, error) {
	file, err := os.Open("config/chatids.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open chatids.json: %w", err)
	}

	defer file.Close() // Ensures file is closed after reading

	var config map[string]ChatIDConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode chatids.json: %w", err)
	}

	// Map public keys to their transaction chat IDs
	chatIDs := make(map[string]string)
	for publicKey, chatConfig := range config {
		chatIDs[publicKey] = chatConfig.Transactions
	}

	return chatIDs, nil
}