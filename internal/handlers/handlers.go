package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/abubakar508/dispatch-ops/internal/geocode"
	"github.com/abubakar508/dispatch-ops/internal/models"
	"github.com/abubakar508/dispatch-ops/internal/optimizer"
	"github.com/abubakar508/dispatch-ops/internal/osrm"
	"github.com/abubakar508/dispatch-ops/internal/validation"
)

type RouteOptimizer interface {
	Optimize(ctx context.Context, req models.RouteRequest) (models.OptimizationResult, error)
}

type Handler struct {
	service  RouteOptimizer
	geocoder geocode.Geocoder
	logger   *slog.Logger
}

func New(service RouteOptimizer, geocoder geocode.Geocoder, logger *slog.Logger) *Handler {
	return &Handler{service: service, geocoder: geocoder, logger: logger}
}

type errorResponse struct {
	Error string `json:"error"`
}

type deliveryRequest struct {
	Points []models.Coordinate `json:"points"`
}

type searchResponse struct {
	Query   string         `json:"query"`
	Results []models.Place `json:"results"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Optimize(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRouteRequest(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "We couldn't read the route you submitted. Please try again.", err)
		return
	}
	h.optimizeAndRespond(w, r, req)
}

func (h *Handler) Deliveries(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload deliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, http.StatusBadRequest, "Send a JSON body with a 'points' array of coordinates.", err)
		return
	}

	req, err := validation.RouteRequestFromPoints(payload.Points)
	if err != nil {
		h.handleOptimizeError(w, err)
		return
	}
	h.optimizeAndRespond(w, r, req)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit := 6
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	results, err := h.geocoder.Search(r.Context(), query, limit)
	if err != nil {
		switch {
		case errors.Is(err, geocode.ErrEmptyQuery):
			h.writeError(w, http.StatusBadRequest, "Type a place to search for.", err)
		case errors.Is(err, geocode.ErrServiceUnavailable):
			h.writeError(w, http.StatusServiceUnavailable, "Location search is temporarily unavailable. Try again shortly.", err)
		default:
			h.writeError(w, http.StatusInternalServerError, "Could not search for that location.", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, searchResponse{Query: query, Results: results})
}

func (h *Handler) optimizeAndRespond(w http.ResponseWriter, r *http.Request, req models.RouteRequest) {
	result, err := h.service.Optimize(r.Context(), req)
	if err != nil {
		h.handleOptimizeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleOptimizeError(w http.ResponseWriter, err error) {
	h.logger.Warn("optimization failed", slog.String("error", err.Error()))

	switch {
	case errors.Is(err, validation.ErrTooFewPoints):
		h.writeError(w, http.StatusUnprocessableEntity,
			"Send at least three coordinates: a start, one delivery, and an end.", err)
	case errors.Is(err, validation.ErrInsufficientStops):
		h.writeError(w, http.StatusUnprocessableEntity,
			"Add a start, an end, and at least one delivery stop before optimizing.", err)
	case errors.Is(err, validation.ErrTooManyStops):
		h.writeError(w, http.StatusUnprocessableEntity,
			"This route has too many stops. Reduce the number of deliveries and try again.", err)
	case errors.Is(err, validation.ErrInvalidLatitude), errors.Is(err, validation.ErrInvalidLongitude):
		h.writeError(w, http.StatusUnprocessableEntity,
			"One of the coordinates is outside the valid range.", err)
	case errors.Is(err, osrm.ErrServiceUnavailable):
		h.writeError(w, http.StatusServiceUnavailable,
			"The routing service is temporarily unavailable. Please try again shortly.", err)
	case errors.Is(err, osrm.ErrNoRoute):
		h.writeError(w, http.StatusUnprocessableEntity,
			"No drivable route connects these points. Check that they are on reachable roads.", err)
	case errors.Is(err, optimizer.ErrTooFewNodes):
		h.writeError(w, http.StatusUnprocessableEntity,
			"Add at least one delivery stop before optimizing.", err)
	case errors.Is(err, context.DeadlineExceeded):
		h.writeError(w, http.StatusGatewayTimeout,
			"Optimization took too long and was cancelled. Try again with fewer stops.", err)
	default:
		h.writeError(w, http.StatusInternalServerError,
			"Something went wrong while optimizing this route.", err)
	}
}

func decodeRouteRequest(r *http.Request) (models.RouteRequest, error) {
	defer r.Body.Close()
	var req models.RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return models.RouteRequest{}, err
	}
	return req, nil
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string, cause error) {
	if cause != nil {
		h.logger.Debug("client-facing error", slog.Int("status", status), slog.String("cause", cause.Error()))
	}
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
