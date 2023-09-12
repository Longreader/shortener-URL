package tools

import (
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const intBytes = "0123456789"

func RandStringBytes(n int) string {
	// randStringBytes - создание короткого URL
	b := make([]byte, n)
	for i := range b {
		if i%2 == 0 {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		} else {
			b[i] = intBytes[rand.Intn(9)]
		}
	}
	return string(b)
}