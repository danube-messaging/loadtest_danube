package workload

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

type PayloadSpec struct {
	SchemaType  string
	MessageSize int
}

// GeneratePayload returns a byte slice appropriate for the schema type.
// MVP: simple synthetic payloads.
func GeneratePayload(spec PayloadSpec, seq uint64) []byte {
	switch strings.ToLower(spec.SchemaType) {
	case "string":
		// Embed only seq header; latency is measured via publish time
		header := fmt.Sprintf("SEQ:%d;", seq)
		sz := spec.MessageSize
		if sz <= 0 {
			sz = 64
		}
		// pad with random content to approximate desired size
		padLen := sz - len(header)
		if padLen < 0 {
			padLen = 0
		}
		return []byte(header + randString(padLen))
	case "json":
		m := map[string]interface{}{
			"seq": seq,
			"msg": randString(16),
		}
		b, _ := json.Marshal(m)
		return b
	case "int64", "number":
		// Send an incrementing number as ASCII bytes
		return []byte(fmt.Sprintf("%d", seq))
	default:
		return []byte("unsupported_schema")
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
