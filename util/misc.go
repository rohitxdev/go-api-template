package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"unicode/utf8"
)

func Ternary[T any](exp bool, trueValue T, falseValue T) T {
	if exp {
		return trueValue
	}
	return falseValue
}

func Coalesce[T any](values ...T) T {
	var res T
	for _, v := range values {
		res = v
		if !reflect.ValueOf(&v).Elem().IsZero() {
			break
		}
	}
	return res
}

func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func ReverseString(s string) string {
	var sb strings.Builder
	runes := []rune(s)
	for i := len(runes) - 1; 0 <= i; i-- {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}

func SanitizeEmail(email string) string {
	emailParts := strings.Split(email, "@")
	username := emailParts[0]
	domain := emailParts[1]
	if strings.Contains(username, "+") {
		username = strings.Split(username, "+")[0]
	}
	username = strings.ReplaceAll(username, "-", "")
	username = strings.ReplaceAll(username, ".", "")
	return username + "@" + domain
}

func PrintTableJSON(jsonData []byte) {
	jsonMap := make(map[string]any)
	json.Unmarshal(jsonData, &jsonMap)
	count := 0
	for key, value := range jsonMap {
		count++
		s := fmt.Sprintf("| %-40.40v | %-40.40v |", key, value)
		fmt.Println(strings.Repeat("-", utf8.RuneCountInString(s)))
		fmt.Println(s)
		if count == len(jsonMap) {
			fmt.Println(strings.Repeat("-", utf8.RuneCountInString(s)))
		}
	}
}

var BufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}
