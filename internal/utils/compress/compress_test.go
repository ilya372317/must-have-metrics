package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressAndDecompress(t *testing.T) {
	tests := []struct {
		name              string
		args              string
		want              string
		wantCompressErr   bool
		wantDecompressErr bool
	}{
		{
			name:              "success case",
			args:              "123",
			want:              "123",
			wantCompressErr:   false,
			wantDecompressErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressedData, cErr := Do([]byte(tt.args))
			if tt.wantCompressErr {
				assert.Error(t, cErr)
				return
			} else {
				require.NoError(t, cErr)
			}
			decompressedData, dErr := Decompress(compressedData)
			if tt.wantDecompressErr {
				assert.Error(t, dErr)
				return
			} else {
				require.NoError(t, dErr)
			}

			assert.Equal(t, tt.want, string(decompressedData))
		})
	}
}
