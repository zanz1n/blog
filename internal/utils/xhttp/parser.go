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

func Parse[T any](req *http.Request) (T, error) {
	value, err := parse[T](req)
	if err != nil {
		err = errutils.NewHttp(err, http.StatusBadRequest, 0, true)
	}

	return value, err
}

func parse[T any](req *http.Request) (T, error) {
	var value T
	ct, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return value, errors.New("parse request: invalid content type")
	}

	switch ct {
	case "application/json":
		if err = json.NewDecoder(req.Body).Decode(&value); err != nil {
			return value, fmt.Errorf("parse request: json: %s", err)
		}

	case "application/x-www-form-urlencoded":
		if err = req.ParseForm(); err != nil {
			return value, fmt.Errorf("parse request: form: %s", err)
		}
		if err = schemaDecoder.Decode(&value, req.PostForm); err != nil {
			return value, fmt.Errorf("parse request: form: %s", err)
		}

	default:
		return value, fmt.Errorf("parse request: invalid content type: %s", ct)
	}

	if err = validate.Struct(&value); err != nil {
		return value, err
	}

	return value, nil
}
