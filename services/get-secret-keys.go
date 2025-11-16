package services

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetSecretKeys() ([]string, error) {
	filePath := "config/secretKeys.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
        return nil, err
    }
	var config struct {
		SecretKeys []string `json:"secretKeys"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.SecretKeys, nil
}

func GetTelegramBotToken() (string, error) {
	filePath := "config/secretKeys.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read secretKeys.json: %w", err)
	}
	var config struct {
		TelegramBotToken string `json:"telegramBotToken"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal secretKeys.json: %w", err)
	}
	if config.TelegramBotToken == "" {
		return "", fmt.Errorf("no Telegram bot token found in secretKeys.json")
	}
	return config.TelegramBotToken, nil
}