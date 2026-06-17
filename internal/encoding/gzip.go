// Package encoding предоставляет утилиты для прозрачного сжатия HTTP-ответов
// и распаковки HTTP-запросов с использованием алгоритма gzip.
// Пакет предназначен для использования в middleware HTTP-серверов.
package encoding

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w              http.ResponseWriter
	zw             *gzip.Writer
	shouldCompress bool
}

// NewCompressWriter создает и возвращает обертку над http.ResponseWriter,
// которая автоматически сжимает ответы в формате gzip.
// Сжатие применяется только для Content-Type, начинающихся с "application/json"
// или "text/html".
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки HTTP-ответа базового http.ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в ответ. Если сжатие включено, данные сначала
// проходят через gzip.Writer.
func (c *compressWriter) Write(p []byte) (int, error) {
	if c.shouldCompress {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

// WriteHeader анализирует заголовок Content-Type. Если он указывает на
// JSON или HTML, включает сжатие и добавляет заголовок Content-Encoding: gzip.
func (c *compressWriter) WriteHeader(statusCode int) {
	ct := c.w.Header().Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.shouldCompress = true
	}

	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer, сбрасывая буферизированные данные в базовый writer.
func (c *compressWriter) Close() error {
	if c.shouldCompress {
		return c.zw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает и возвращает обертку над io.ReadCloser для чтения
// входящего тела запроса, сжатого с помощью gzip.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает распакованные данные из базового gzip-ридера в предоставленный буфер.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает как gzip-ридер, так и базовый поток чтения (io.ReadCloser).
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
