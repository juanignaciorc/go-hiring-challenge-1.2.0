package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/mytheresa/go-hiring-challenge/app/errs"
	"github.com/mytheresa/go-hiring-challenge/app/logz"
)

// AppHandler is a handler that can return an error handled centrally.
type AppHandler func(w http.ResponseWriter, r *http.Request) error

// Serve executes an AppHandler applying centralized recovery, logging, and error serialization.
func Serve(w http.ResponseWriter, r *http.Request, h AppHandler) {
	start := time.Now()

	// Inject request ID and logger into context
	reqID := requestID(r)
	lg := logz.New().With(logz.Fields{"request_id": reqID, "path": r.URL.Path, "method": r.Method})
	r = r.WithContext(logz.IntoContext(logz.WithRequestID(r.Context(), reqID), lg))

	// Ensure JSON content type on errors
	w.Header().Set("Content-Type", "application/json")

	// Panic recovery
	defer func() {
		if rec := recover(); rec != nil {
			lg.Error("panic recovered", logz.Fields{"panic": rec, "stack": string(debug.Stack())})
			writeAppError(w, errs.Internal("internal server error"))
		}
	}()

	err := h(w, r)

	status := httpStatusFromWriter(w)
	if err != nil {
		// Centralized error handling
		ae := errs.From(err)
		lg.Error("request failed", logz.Fields{"code": ae.Code, "error": ae.Error()})
		writeAppError(w, ae)
		status = errs.HTTPStatus(ae.Code)
	}

	lg.Info("request completed", logz.Fields{"status": status, "duration_ms": time.Since(start).Milliseconds()})
}

// Wrap converts an AppHandler into a standard http.Handler with centralized
// panic recovery, structured logging, and error serialization.
func Wrap(h AppHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Serve(w, r, h)
	})
}

// writeAppError writes a consistent error payload and status code.
func writeAppError(w http.ResponseWriter, e *errs.AppError) {
	status := errs.HTTPStatus(e.Code)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string    `json:"error"`
		Code  errs.Code `json:"code"`
	}{Error: e.Message, Code: e.Code})
}

// httpStatusFromWriter attempts to fetch the current status if the writer implements interface.
// For simplicity, if not available we'll return 200.
func httpStatusFromWriter(w http.ResponseWriter) int { return 200 }

// requestID returns an ID for correlating logs. If the incoming request has a standard header
// it will be honored; otherwise a timestamp-based ID is generated.
func requestID(r *http.Request) string {
	if v := r.Header.Get("X-Request-ID"); v != "" {
		return v
	}
	return time.Now().UTC().Format("20060102T150405.000000000Z07:00")
}
