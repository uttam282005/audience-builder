package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mable/audience-builder/backend/internal/evaluator"
	"github.com/mable/audience-builder/backend/internal/model"
	"github.com/mable/audience-builder/backend/internal/validator"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	logger *slog.Logger
	eval   *evaluator.Evaluator
}

// NewServer constructs the HTTP server and router.
func NewServer(logger *slog.Logger, eval *evaluator.Evaluator) http.Handler {
	s := &Server{
		logger: logger,
		eval:   eval,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /v1/audiences/preview", s.handlePreview)

	// Wrap middleware chain: Recovery -> Logging -> CORS -> Mux
	var handler http.Handler = mux
	handler = CORSMiddleware(handler)
	handler = LoggingMiddleware(logger)(handler)
	handler = RecoveryMiddleware(logger)(handler)

	return handler
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	// 1. Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// 2. Decode JSON with strict unknown field rejection
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req model.PreviewRequest
	if err := dec.Decode(&req); err != nil {
		var fieldErrors []model.FieldError
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			fieldErrors = append(fieldErrors, model.FieldError{
				Field: "body",
				Issue: "malformed JSON syntax",
			})
		case errors.As(err, &unmarshalTypeErr):
			fieldErrors = append(fieldErrors, model.FieldError{
				Field: unmarshalTypeErr.Field,
				Issue: "incorrect data type provided",
			})
		case errors.Is(err, io.EOF):
			fieldErrors = append(fieldErrors, model.FieldError{
				Field: "body",
				Issue: "request body cannot be empty",
			})
		default:
			fieldErrors = append(fieldErrors, model.FieldError{
				Field: "body",
				Issue: err.Error(),
			})
		}

		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_json",
			Message: "Failed to parse JSON request body",
			Details: fieldErrors,
		})
		return
	}

	// 3. Domain validation
	if valErrors := validator.ValidatePreviewRequest(&req); len(valErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid audience rule definition",
			Details: valErrors,
		})
		return
	}

	// 4. Rule Evaluation
	resp, err := s.eval.Evaluate(r.Context(), &req)
	if err != nil {
		s.logger.Error("audience_evaluation_failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error:   "evaluation_error",
			Message: "An error occurred while evaluating audience conditions",
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
