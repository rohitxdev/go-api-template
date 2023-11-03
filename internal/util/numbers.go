package util

import "math/rand"

func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
