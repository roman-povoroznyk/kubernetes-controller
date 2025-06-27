package utils

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
)

// ResponseHelper provides common HTTP response utilities
type ResponseHelper struct{}

// MethodNotAllowed sends a 405 Method Not Allowed response
func (r *ResponseHelper) MethodNotAllowed(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
	ctx.SetContentType("application/json")
	response := map[string]string{
		"error":   "Method Not Allowed",
		"message": fmt.Sprintf("Method %s is not allowed for this endpoint", ctx.Method()),
	}
	json.NewEncoder(ctx).Encode(response)
}

// NotFound sends a 404 Not Found response
func (r *ResponseHelper) NotFound(ctx *fasthttp.RequestCtx, resource string) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	response := map[string]string{
		"error":   "Not Found",
		"message": fmt.Sprintf("%s not found", resource),
	}
	json.NewEncoder(ctx).Encode(response)
}

// InternalServerError sends a 500 Internal Server Error response
func (r *ResponseHelper) InternalServerError(ctx *fasthttp.RequestCtx, err error) {
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetContentType("application/json")
	response := map[string]string{
		"error":   "Internal Server Error",
		"message": err.Error(),
	}
	json.NewEncoder(ctx).Encode(response)
}

// JSONResponse sends a JSON response with given status code
func (r *ResponseHelper) JSONResponse(ctx *fasthttp.RequestCtx, statusCode int, data interface{}) error {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	return json.NewEncoder(ctx).Encode(data)
}

// PlainResponse sends a plain text response
func (r *ResponseHelper) PlainResponse(ctx *fasthttp.RequestCtx, statusCode int, message string) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("text/plain")
	fmt.Fprint(ctx, message)
}

// Global instance for easy access
var HTTP = &ResponseHelper{}
