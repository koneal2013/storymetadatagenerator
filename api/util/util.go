package util

import (
	"os"
	"strings"

	"go.uber.org/zap"
)

func GetEasyWords() (map[string]struct{}, error) {
	aw := map[string]struct{}{}

	easyWords, err := os.ReadFile("./api/util/easy_words.txt")
	if err != nil {
		zap.L().Sugar().Error(err)
		return nil, err
	}

	words := strings.Fields(string(easyWords))

	for _, w := range words {
		aw[w] = struct{}{}
	}
	return aw, nil
}
