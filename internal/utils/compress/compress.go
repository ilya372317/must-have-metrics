package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

type Writer struct {
	w          http.ResponseWriter
	gzipWriter *gzip.Writer
}

func NewWriter(writer http.ResponseWriter) *Writer {
	gzipWriter := gzip.NewWriter(writer)

	return &Writer{
		w:          writer,
		gzipWriter: gzipWriter,
	}
}

func (w *Writer) Header() http.Header {
	return w.w.Header()
}

func (w *Writer) Write(bytes []byte) (int, error) {
	return w.gzipWriter.Write(bytes)
}

func (w *Writer) WriteHeader(statusCode int) {
	if statusCode < 300 {
		w.w.Header().Set("Content-Encoding", "gzip")
	}

	w.w.WriteHeader(statusCode)
}

func (w *Writer) Close() error {
	return w.gzipWriter.Close()
}

type Reader struct {
	r          io.ReadCloser
	gzipReader *gzip.Reader
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return r.gzipReader.Read(p)
}

func (r *Reader) Close() error {
	err := r.r.Close()
	if err != nil {
		return err
	}

	return r.gzipReader.Close()
}

func NewReader(reader io.ReadCloser) (*Reader, error) {
	gReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	return &Reader{
		r:          reader,
		gzipReader: gReader,
	}, nil
}

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

func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed create gzip reader: %w", err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
