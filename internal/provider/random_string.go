package provider

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	numeric = []rune("0123456789")
	lower   = []rune("abcdefghijklmnopqrstuvwxyz")
	upper   = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func generateRandomString(length int, includeNumeric bool, includeLowercase bool, includeUppercase bool) string {
	setToUse := make([]rune, 0)
	if includeNumeric {
		setToUse = append(setToUse, numeric...)
	}
	if includeLowercase {
		setToUse = append(setToUse, lower...)
	}
	if includeUppercase {
		setToUse = append(setToUse, upper...)
	}

	b := make([]rune, length)
	for i := range b {
		b[i] = setToUse[rand.Intn(len(setToUse))]
	}
	return string(b)
}
