package main

import (
	"crypto/rand"
	"math/big"
)

const (
	lower       = "abcdefghijklmnopqrstuvwxyz"
	upper       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits      = "0123456789"
	specials    = "!@#%^_=+"
	allChars    = lower + upper + digits + specials
	passwordLen = 16
)

func getRandomChar(charset string) byte {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	return charset[n.Int64()]
}

func shuffle(bytes []byte) {
	for i := len(bytes) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		bytes[i], bytes[j.Int64()] = bytes[j.Int64()], bytes[i]
	}
}

func generatePassword(length int) string {
	if length < 4 {
		panic("Password length must be at least 4 to include all character types.")
	}

	// Ensure at least one character from each required set
	password := []byte{
		getRandomChar(lower),
		getRandomChar(upper),
		getRandomChar(digits),
		getRandomChar(specials),
	}

	// Fill the rest with random characters from all sets
	for i := 4; i < length; i++ {
		password = append(password, getRandomChar(allChars))
	}

	// Shuffle the final password
	shuffle(password)

	return string(password)
}

func generateNickname() string {
	nickname := make([]byte, 8)
	for i := range nickname {
		nickname[i] = getRandomChar(lower + upper + digits)
	}
	return string(nickname)
}