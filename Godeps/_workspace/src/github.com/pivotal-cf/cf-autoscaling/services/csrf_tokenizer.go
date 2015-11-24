package services

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"time"
)

type CSRFTokenizer struct{}

func (t CSRFTokenizer) Generate() string {
	salt, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic(err)
	}

	now := time.Now().Unix()
	hash := sha1.New()
	io.WriteString(hash, fmt.Sprintf("%d %d", salt, now))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
