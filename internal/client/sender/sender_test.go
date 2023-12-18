package sender

import (
	"crypto/hmac"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_createSign(t *testing.T) {
	type argument struct {
		src,
		key string
	}
	tests := []struct {
		name string
		arg  argument
	}{
		{
			name: "simple success case",
			arg: argument{
				src: "Ilya Otinov",
				key: "1234567",
			},
		},
		{
			name: "long success case",
			arg: argument{
				src: "It is very long and complex string, and i expect hash to be equal anyway",
				key: "my-name-is-ilya",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createSign([]byte(tt.arg.src), tt.arg.key)

			h := hmac.New(sha256.New, []byte(tt.arg.key))
			h.Write([]byte(tt.arg.src))
			assert.Equal(t, h.Sum(nil), got)
		})
	}
}
