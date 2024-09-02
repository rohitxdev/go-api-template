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

// Prints given data in table format. Accepts structs, marshalled JSON or []byte
func PrintTableJSON(v any) {
	jsonMap := make(map[string]any)
	if b, ok := v.([]byte); ok {
		if err := json.Unmarshal(b, &jsonMap); err != nil {
			fmt.Println(err)
		}
	} else {
		jsonData, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
			return
		}
		_ = json.Unmarshal(jsonData, &jsonMap)
	}

	count := 0
	for key, value := range jsonMap {
		count++
		data := value
		if reflect.TypeOf(value).Kind() == reflect.Map {
			data = "__invalid_type__"
		}
		s := fmt.Sprintf("| %-40.40v | %-40.40v |", key, data)
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
