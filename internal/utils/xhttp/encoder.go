package xhttp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/elnormous/contenttype"
	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/web/templates"
)

var (
	ctypeHtml = contenttype.NewMediaType("text/html")
	ctypeJson = contenttype.NewMediaType("application/json")
)

var mediaTypes = []contenttype.MediaType{ctypeHtml, ctypeJson}

type ComponentFunc[T any] func(T) templ.Component

func Component[T any](
	c *Ctx,
	cf ComponentFunc[T],
	v T,
	code int,
) error {
	return handler(c, cf, v, code, false)
}

func Error(c *Ctx, p templates.PageData[error]) {
	errd := errutils.Http(p.Data)
	data := templates.ErrorData{
		Code:       errd.ErrorCode(),
		HttpStatus: errd.HttpStatus(),
	}

	if errd.Transparent() {
		data.Message = http.StatusText(data.HttpStatus)
	} else {
		data.Message = errd.Error()
	}

	p2 := templates.PageData[templates.ErrorData]{
		Name:  p.Name,
		Token: p.Token,
		Data:  data,
	}

	handler(c, templates.ErrorPage, p2, data.HttpStatus, true)
}

func handler[T any](
	c *Ctx,
	cf ComponentFunc[T],
	v T,
	code int,
	ignoreparsing bool,
) error {
	mt, _, err := contenttype.GetAcceptableMediaType(c.Request, mediaTypes)
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
		if err = encodeJson(c, v, code); err != nil {
			return fmt.Errorf("encode json response: %s", err)
		}
	} else {
		comp := cf(v)
		err = encodeTemplate(comp, c, c.Request, code)
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
