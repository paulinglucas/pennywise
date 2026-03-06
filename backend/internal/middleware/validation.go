package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
)

func Validation(specBytes []byte, basePath string) (func(http.Handler) http.Handler, error) {
	specRouter, err := loadSpec(specBytes, basePath)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := specRouter.FindRoute(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			input := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
				},
			}

			if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
				writeValidationError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

func loadSpec(specBytes []byte, basePath string) (routers.Router, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(specBytes)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(context.Background()); err != nil {
		return nil, err
	}
	doc.Servers = []*openapi3.Server{{URL: basePath}}
	return legacy.NewRouter(doc)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, validationErr error) {
	message := validationErr.Error()
	if idx := strings.Index(message, "\n"); idx > 0 {
		message = message[:idx]
	}
	if len(message) > 200 {
		message = message[:200]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	resp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":       "VALIDATION_FAILED",
			"message":    message,
			"request_id": GetRequestID(r.Context()),
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}
