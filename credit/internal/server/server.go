package server

import (
	"credit/internal/manager"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler exposes HTTP endpoints for credit operations
type Handler struct {
	mgr *manager.Manager
}

func NewHandler(m *manager.Manager) *Handler {
	return &Handler{mgr: m}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/debit", h.handleDebit)
	r.Post("/credit", h.handleCredit)
	r.Get("/balance/{org}", h.handleBalance)
	return r
}

func NewServer(addr string, h *Handler) *http.Server {
	handler := h.Routes()
	srv := &http.Server{Addr: addr, Handler: handler}
	return srv
}

// request/response types and handlers
type debitReq struct {
	OrgID  string `json:"org_id"`
	Amount int    `json:"amount"`
}

type creditReq struct {
	OrgID  string `json:"org_id"`
	Amount int    `json:"amount"`
}

func (h *Handler) handleDebit(w http.ResponseWriter, r *http.Request) {
	var req debitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.OrgID == "" || req.Amount <= 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	rem, err := h.mgr.Debit(r.Context(), req.OrgID, req.Amount, "")
	if errors.Is(err, manager.ErrInsufficient) {
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
		return
	} else if err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]int{"remaining": rem})
}

func (h *Handler) handleCredit(w http.ResponseWriter, r *http.Request) {
	var req creditReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.OrgID == "" || req.Amount <= 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	res, err := h.mgr.Credit(r.Context(), req.OrgID, req.Amount, "")
	if err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]int{"balance": res})
}

func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request) {
	org := chi.URLParam(r, "org")
	if org == "" {
		http.Error(w, "missing org", http.StatusBadRequest)
		return
	}
	b := h.mgr.Balance(r.Context(), org)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"balance": b})
}
