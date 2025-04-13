package utils

import (
	"crypto/rand"
	"errors"
	mrand "math/rand/v2"
)

var Alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int, chars []rune) string {
	if len(chars) == 0 {
		panic(errors.New("the provided chars array is empty"))
	}

	b := make([]rune, n)
	for i := range b {
		b[i] = chars[mrand.IntN(len(chars))]
	}
	return string(b)
}

func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
