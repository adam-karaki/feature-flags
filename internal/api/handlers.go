package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"feature-flags/internal/flags"
)

type Handler struct{ service *flags.Service }

func NewHandler(service *flags.Service) *Handler { return &Handler{service: service} }

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/flags", h.flags)
	mux.HandleFunc("/flags/", h.flag)
	return logging(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) flags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var f flags.Flag
	if err := decode(r, &f); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.Create(r.Context(), &f); err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *Handler) flag(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/flags/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	name := parts[0]
	if len(parts) == 2 && parts[1] == "evaluate" {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var input flags.EvaluationContext
		if err := decode(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result, err := h.service.Evaluate(r.Context(), name, input)
		if err != nil {
			handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		f, err := h.service.Get(r.Context(), name)
		if err != nil {
			handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, f)
	case http.MethodPut:
		version := r.Header.Get("If-Match")
		if version == "" {
			writeError(w, http.StatusBadRequest, errors.New("If-Match header is required"))
			return
		}
		expected, err := flags.ParseVersion(version)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("If-Match must be an integer version"))
			return
		}
		var f flags.Flag
		if err := decode(r, &f); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		f.Name = name
		if err := h.service.Update(r.Context(), &f, expected); err != nil {
			handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, f)
	case http.MethodDelete:
		version := r.Header.Get("If-Match")
		expected, err := flags.ParseVersion(version)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("If-Match must be an integer version"))
			return
		}
		if err := h.service.Delete(r.Context(), name, expected); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w)
	}
}

func decode(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST, PUT, DELETE")
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, flags.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, flags.ErrConflict):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
