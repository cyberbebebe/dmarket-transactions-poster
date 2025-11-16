package services

import (
	"fmt"
	"time"
)

func MakeKeysStamps(secretKeys []string) map[string]string {
	keysStamps := make(map[string]string)
	for _, key := range secretKeys {
		keysStamps[key] = fmt.Sprintf("%d", time.Now().Unix())
	}
	return keysStamps
}