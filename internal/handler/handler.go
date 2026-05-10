package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"subscription-service/internal/model"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

type Handler struct {
	svc *service.SubscriptionService
	log *zap.Logger
}

func New(svc *service.SubscriptionService, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"}}))
	r.Get("/health", h.health)
	r.Get("/swagger/openapi.yaml", h.swagger)
	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/sum", h.sum)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})
	return r
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (h *Handler) swagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "docs/openapi.yaml")
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	sub, err := h.svc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.log.Info("subscription created", zap.Int64("id", sub.ID))
	writeJSON(w, http.StatusCreated, model.ToResponse(sub))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	sub, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		h.log.Error("get failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, model.ToResponse(sub))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	uid, err := repository.ParseUUIDPtr(q.Get("user_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	var serviceName *string
	if v := strings.TrimSpace(q.Get("service_name")); v != "" {
		serviceName = &v
	}
	items, err := h.svc.List(r.Context(), model.ListFilter{Limit: limit, Offset: offset, UserID: uid, ServiceName: serviceName})
	if err != nil {
		h.log.Error("list failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	resp := make([]model.SubscriptionResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, model.ToResponse(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	sub, err := h.svc.Update(r.Context(), id, req)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.log.Info("subscription updated", zap.Int64("id", sub.ID))
	writeJSON(w, http.StatusOK, model.ToResponse(sub))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	} else if err != nil {
		h.log.Error("delete failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	h.log.Info("subscription deleted", zap.Int64("id", id))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sum(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, err := model.ParseMonth(q.Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from, expected MM-YYYY")
		return
	}
	to, err := model.ParseMonth(q.Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to, expected MM-YYYY")
		return
	}
	uid, err := repository.ParseUUIDPtr(q.Get("user_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	var serviceName *string
	if v := strings.TrimSpace(q.Get("service_name")); v != "" {
		serviceName = &v
	}
	filter := model.SumFilter{From: repository.MonthStart(from), To: repository.MonthStart(to), UserID: uid, ServiceName: serviceName}
	total, err := h.svc.Sum(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := model.SumResponse{Total: total, From: model.FormatMonth(filter.From), To: model.FormatMonth(filter.To)}
	if uid != nil {
		resp.UserID = uid.String()
	}
	if serviceName != nil {
		resp.ServiceName = *serviceName
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}
