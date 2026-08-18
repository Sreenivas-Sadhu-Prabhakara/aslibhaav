package backend

import (
	"encoding/json"
	"net/http"
	"time"
)

// EvaluateResponse is returned by POST /evaluate.
type EvaluateResponse struct {
	EffectiveCostPerUnit float64   `json:"effectiveCostPerUnit"`
	Breakdown            Breakdown `json:"breakdown"`
}

// Evaluation is one saved scheme evaluation (for /history).
type Evaluation struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Input     Input     `json:"input"`
	Effective float64   `json:"effective"`
	Note      string    `json:"note,omitempty"`
}

// HistoryStore persists and lists past evaluations.
type HistoryStore interface {
	Save(ev Evaluation) (Evaluation, error)
	List(limit int) ([]Evaluation, error)
}

// NewServer wires the routes. store may be nil when no database is configured;
// /evaluate still works, and /history reports that history is unavailable.
func NewServer(store HistoryStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/evaluate", evaluateHandler(store))
	mux.HandleFunc("/history", historyHandler(store))
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func evaluateHandler(store HistoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var in Input
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := in.Validate(); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		eff, bd := EffectiveCostPerUnit(in)
		if store != nil {
			_, _ = store.Save(Evaluation{CreatedAt: time.Now().UTC(), Input: in, Effective: eff})
		}
		writeJSON(w, http.StatusOK, EvaluateResponse{EffectiveCostPerUnit: eff, Breakdown: bd})
	}
}

func historyHandler(store HistoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeErr(w, http.StatusServiceUnavailable, "history unavailable: no database configured")
			return
		}
		items, err := store.List(50)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "could not read history")
			return
		}
		writeJSON(w, http.StatusOK, items)
	}
}
