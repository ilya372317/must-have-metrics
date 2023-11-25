package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

const LastPositiveStatusCode = 300

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
	size, err := w.gzipWriter.Write(bytes)
	if err != nil {
		err = fmt.Errorf("failed compress data: %w", err)
	}
	return size, err
}

func (w *Writer) WriteHeader(statusCode int) {
	if statusCode < LastPositiveStatusCode {
		w.w.Header().Set("Content-Encoding", "gzip")
	}

	w.w.WriteHeader(statusCode)
}

func (w *Writer) Close() error {
	err := w.gzipWriter.Close()
	if err != nil {
		err = fmt.Errorf("failed close gzip response writer: %w", err)
	}
	return err
}

type Reader struct {
	r          io.ReadCloser
	gzipReader *gzip.Reader
}

func (r *Reader) Read(p []byte) (n int, err error) {
	size, err := r.gzipReader.Read(p)
	if err != nil {
		err = fmt.Errorf("failed read from gzip response reader: %w", err)
	}
	return size, err
}

func (r *Reader) Close() error {
	err := r.r.Close()
	if err != nil {
		return fmt.Errorf("failed close response reader: %w", err)
	}

	err = r.gzipReader.Close()
	if err != nil {
		err = fmt.Errorf("failed close gzip reader: %w", err)
	}

	return err
}

func NewReader(reader io.ReadCloser) (*Reader, error) {
	gReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed create new gzip reader: %w", err)
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
