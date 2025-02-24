package xhttp

import "net/http"

func NewWriterWrapper(w http.ResponseWriter) *WriterWrapper {
	return &WriterWrapper{
		ResponseWriter: w,
	}
}

type WriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *WriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *WriterWrapper) GetStatus() int {
	return w.statusCode
}
