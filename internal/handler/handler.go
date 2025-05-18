package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dsha256/dispatcher/internal/dispatcher"
	"github.com/dsha256/dispatcher/internal/middleware"
	"github.com/dsha256/dispatcher/internal/responder"
)

var ErrMethodNotAllowed = errors.New("method not allowed")

type Handler struct {
	logger     *slog.Logger
	dispatcher *dispatcher.Dispatcher
}

func New(
	logger *slog.Logger,
	dispatcher *dispatcher.Dispatcher,
) *Handler {
	return &Handler{
		logger:     logger,
		dispatcher: dispatcher,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/dispatcher/itinerary", h.wrapHandler(h.handleItinerary))
	mux.Handle("/api/v1/liveness", h.wrapHandler(h.handleLiveness))
	mux.Handle("/api/v1/readiness", h.wrapHandler(h.handleReadiness))
	h.logger.Info("Routes registered")
}

func (h *Handler) wrapHandler(handler http.HandlerFunc) http.Handler {
	return middleware.LoggingMiddleware(
		h.logger,
		middleware.RecoveryMiddleware(
			h.logger,
			handler,
		),
	)
}

func (h *Handler) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	responder.WriteSuccess(w, http.StatusOK, "All services are up and running", json.RawMessage{})
}

func (h *Handler) handleReadiness(w http.ResponseWriter, _ *http.Request) {
	responder.WriteSuccess(w, http.StatusOK, "All services are up and ready to process requests", json.RawMessage{})
}

func (h *Handler) handleError(w http.ResponseWriter, err error, status int) {
	h.logger.Error("Error handling request", "error", err)
	responder.WriteError(w, status, err)
}

func (h *Handler) isBadRequestError(err error) bool {
	return errors.Is(err, dispatcher.ErrDifferentStartingPoints) ||
		errors.Is(err, dispatcher.ErrMultipleSameDestination) ||
		errors.Is(err, dispatcher.ErrCycleInItinerary)
}
