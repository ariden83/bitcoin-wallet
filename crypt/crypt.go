package crypt

import (
	"fmt"
	"crypto/rand"
	"encoding/hex"
)

// CreateHash returns random 8 byte array as a hex string
func CreateHash() string {
	key := make([]byte, 8)

	_, err := rand.Read(key)
	if err != nil {
		// handle error here
		fmt.Println(err)
	}

	str := hex.EncodeToString(key)

	return str
}
