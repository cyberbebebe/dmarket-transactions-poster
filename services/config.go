package services

import (
	"encoding/json"
	"os"

	"github.com/cyberbebebe/dmarket-transactions-poster/types"
)

func LoadConfig(path string) ([]types.AccountConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var configs []types.AccountConfig
	if err := json.NewDecoder(file).Decode(&configs); err != nil {
		return nil, err
	}
	return configs, nil
}