package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zanz1n/blog/internal/utils/errutils"
)

func marshalUnsafe(code int, error string) []byte {
	return []byte(fmt.Sprintf(`{"code":%d,"error":"%s"}`, code, error))
}

type httpErrorData struct {
	Code  int32  `json:"code"`
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, err error) {
	e := errutils.Http(err)
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(e.HttpStatus())

	data := httpErrorData{
		Code: e.ErrorCode(),
	}

	if !e.Transparent() {
		data.Error = http.StatusText(e.HttpStatus())
	} else {
		data.Error = e.Error()
	}

	b, e2 := json.Marshal(data)
	if e2 != nil {
		st := http.StatusInternalServerError
		b = marshalUnsafe(st, http.StatusText(st))
	}

	_, _ = w.Write(b)
}
