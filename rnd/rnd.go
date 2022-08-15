package rnd

import (
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

// RandomStr generates random string with provided length
func RandomStr(length int) string {
	if length <= 0 {
		return ""
	}

	buff := make([]rune, length)

	for i := range buff {
		buff[i] = letterRunes[RandomInt(len(letterRunes))]
	}

	return string(buff)
}

// RandomInt generates random int, max should be > 0
func RandomInt(max int) int {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	return r.Intn(max)
}
