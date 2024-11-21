package binding

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// DecodeJSONBody decodes the http request body into the given struct and validates it
func DecodeJSONBody(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return err
	}

	// validate the decode body
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	err := validate.Struct(v)
	if err != nil {
		return handleBindingErrors(err)
	}

	return nil
}
