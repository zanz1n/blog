package xhttp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/elnormous/contenttype"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

var (
	ctypeHtml = contenttype.NewMediaType("text/html")
	ctypeJson = contenttype.NewMediaType("application/json")
)

var mediaTypes = []contenttype.MediaType{ctypeHtml, ctypeJson}

type ComponentFunc[T any] func(T) templ.Component

type ComponentWriter[T any] struct {
	w http.ResponseWriter
	r *http.Request
	c ComponentFunc[T]
}

func Component[T any](
	w http.ResponseWriter,
	r *http.Request,
	cf ComponentFunc[T],
	v T,
	code int,
) error {
	return handler(w, r, cf, v, code, false)
}

func handler[T any](
	w http.ResponseWriter,
	r *http.Request,
	cf ComponentFunc[T],
	v T,
	code int,
	ignoreparsing bool,
) error {
	mt, _, err := contenttype.GetAcceptableMediaType(r, mediaTypes)
	if err != nil && !ignoreparsing {
		return errutils.NewHttp(
			fmt.Errorf("content negotiation: parse Accept header: %s", err),
			http.StatusBadRequest,
			0,
			true,
		)
	}

	isWildcard := mt.Type == `*` && mt.Subtype == `*`
	if !isWildcard && mt.Matches(ctypeJson) {
		if err = encodeJson(w, v, code); err != nil {
			return fmt.Errorf("encode json response: %s", err)
		}
	} else {
		comp := cf(v)
		err = encodeTemplate(comp, w, r, code)
		if err != nil {
			return fmt.Errorf("encode html response: %s", err)
		}
	}

	return nil
}

func encodeTemplate(
	c templ.Component,
	w http.ResponseWriter,
	req *http.Request,
	status int,
) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	err := c.Render(req.Context(), buf)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
	return nil
}

func encodeJson(w http.ResponseWriter, v any, status int) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf)

	return nil
}
