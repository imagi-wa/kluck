package auth

import (
	"math/rand"
	"time"
)

var randSrc = rand.NewSource(time.Now().UnixNano())

const (
	letters         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIndexBits = 5
	letterIndexMask = 1<<letterIndexBits - 1
	letterIndexMax  = 63 / letterIndexBits
)

func GenerateRandomBytes(length int) []byte {
	key := make([]byte, length)
	cache, remain := randSrc.Int63(), letterIndexMax
	for i := length - 1; i >= 0; {
		if remain == 0 {
			cache, remain = randSrc.Int63(), letterIndexMax
		}
		index := int(cache & letterIndexMask)
		if index < len(letters) {
			key[i] = letters[index]
			i--
		}
		cache >>= letterIndexBits
		remain--
	}
	return key
}

func GenerateRandomString(length int) string {
	str := string(GenerateRandomBytes(length))
	return str
}
