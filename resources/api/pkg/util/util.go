package util

import (
	"net/http"
	"net/http/httputil"

	"api/pkg/log"

	"go.uber.org/zap"
)

// RequestDump is
func RequestDump(r *http.Request) {

	logger := log.LoggerWithLambdaRqID(r.Context())

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.Error("unable to dump request", zap.Error(err))
	}

	logger.Debug(string(requestDump))

	return
}
