package service

import (
	"math/rand"
	"time"
)

// RandomDelay returns a random duration between min and max milliseconds.
func RandomDelay(minMs, maxMs int) time.Duration {
	if maxMs <= minMs {
		return time.Duration(minMs) * time.Millisecond
	}
	ms := minMs + rand.Intn(maxMs-minMs)
	return time.Duration(ms) * time.Millisecond
}

// RandomBetween returns a random int between min and max (inclusive).
func RandomBetween(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}
