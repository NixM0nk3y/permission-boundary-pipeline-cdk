package api

import (
	"net/http"

	"api/pkg/log"
	"api/pkg/util"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

type helloworldResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (u *helloworldResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {

	logger := log.LoggerWithLambdaRqID(r.Context())

	logger.Debug("EventHandler")

	util.RequestDump(r)

	render.Render(w, r, &helloworldResponse{
		Status:  200,
		Message: "hello world",
	})
}
