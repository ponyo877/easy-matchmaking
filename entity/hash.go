package entity

import (
	"crypto/sha256"
	"fmt"
	"time"
)

var (
	headCount = 7
)

type Hash struct {
	seed time.Time
}

func NewHash(seed time.Time) *Hash {
	return &Hash{seed}
}

func (h *Hash) String() string {
	sha := sha256.New()
	if _, err := sha.Write([]byte(h.seed.String())); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha.Sum(nil))[:headCount]
}
