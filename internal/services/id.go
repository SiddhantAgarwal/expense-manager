package services

import (
	"crypto/rand"
	"fmt"
)

func NewID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	return fmt.Sprintf("%x", b)
}
