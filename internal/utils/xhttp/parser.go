package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

var validate = validator.New()
var schemaDecoder = schema.NewDecoder()

func init() {
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		s := field.Tag.Get("json")
		s, _, _ = strings.Cut(s, ",")
		return s
	})
	schemaDecoder.SetAliasTag("json")
}

func IsHtmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func Parse[T any](req *http.Request) (T, error) {
	var value T
	ct, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return value, errutils.NewHttp(
			errors.New("parse request: invalid content type"),
			http.StatusBadRequest,
			http.StatusBadRequest,
			true,
		)
	}

	switch ct {
	case "application/json":
		if err = json.NewDecoder(req.Body).Decode(&value); err != nil {
			return value, errutils.NewHttp(
				fmt.Errorf("parse request: json: %s", err),
				http.StatusUnprocessableEntity,
				http.StatusUnprocessableEntity,
				true,
			)
		}

	case "application/x-www-form-urlencoded":
		if err = parseFormReq(&value, req); err != nil {
			return value, errutils.NewHttp(
				fmt.Errorf("parse request: form: %s", err),
				http.StatusUnprocessableEntity,
				http.StatusUnprocessableEntity,
				true,
			)
		}

	default:
		return value, errutils.NewHttp(
			fmt.Errorf("parse request: invalid content type: %s", ct),
			http.StatusBadRequest,
			http.StatusBadRequest,
			true,
		)
	}

	if err = validate.Struct(&value); err != nil {
		return value, convertValidateError(err)
	}

	return value, nil
}

func parseFormReq(dst any, req *http.Request) (err error) {
	if err = req.ParseForm(); err != nil {
		return
	}
	err = schemaDecoder.Decode(dst, req.PostForm)
	return
}

func convertValidateError(err error) error {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var b strings.Builder

	for _, err := range errs {
		b.WriteString(fmt.Sprintf(
			"Field %s invalid: %s criteria\n",
			err.Field(),
			err.Tag(),
		))
	}

	return errutils.NewHttpS(
		b.String(),
		http.StatusUnprocessableEntity,
		http.StatusUnprocessableEntity,
		true,
	)
}
