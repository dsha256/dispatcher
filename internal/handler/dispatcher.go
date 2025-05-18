package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dsha256/dispatcher/internal/responder"
)

func (h *Handler) handleItinerary(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.reconstructItinerary(w, r)
	default:
		h.handleError(w, ErrMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

type ReconstructItineraryRequest struct {
	Tickets [][]string `json:"tickets"`
}

func (h *Handler) reconstructItinerary(w http.ResponseWriter, r *http.Request) {
	var req ReconstructItineraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "error decoding request body", "error", err, "payload", req, "path", r.URL.Path)
		h.handleError(w, err, http.StatusBadRequest)

		return
	}

	linearPath, err := h.dispatcher.ReconstructItinerary(r.Context(), &req.Tickets)
	if err != nil {
		if h.isBadRequestError(err) {
			h.logger.WarnContext(r.Context(), "error calculating linear path", "error", err, "payload", req, "path", r.URL.Path)
			h.handleError(w, err, http.StatusBadRequest)

			return
		}
		h.logger.ErrorContext(r.Context(), "error calculating linear path", "error", err)
		h.handleError(w, err, http.StatusInternalServerError)

		return
	}

	responder.WriteSuccess(w, http.StatusOK, "", map[string][]string{
		"linear_path": linearPath,
	})
}
