package utils

import (
	"math/rand"
	"strings"
)

var RandDict_Default = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func GetRandomString(length int, dict []rune) string {
	if dict == nil {
		dict = RandDict_Default
	}
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteRune(dict[rand.Intn(len(dict))])
	}
	return sb.String()
}

func Contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if needle == item {
			return true
		}
	}
	return false
}
