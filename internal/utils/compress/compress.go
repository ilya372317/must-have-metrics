package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

// Do compress given data by gzip algo.
func Do(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed create gzip writer: %w", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed close gzip writer: %w", err)
	}
	return b.Bytes(), nil
}

// Decompress decompres given data by gzip algo.
func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed create gzip reader: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %w", err)
	}

	return b.Bytes(), nil
}
