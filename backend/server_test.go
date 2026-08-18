package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEvaluateEndpoint_OK(t *testing.T) {
	srv := NewServer(nil) // no DB needed for /evaluate
	body := `{"basePrice":100,"tradeDiscountPct":10,"cashDiscountPct":5,"freeBuyQty":10,"freeGetQty":2,"freight":120,"freightExtra":true}`
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var resp EvaluateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !almost(resp.EffectiveCostPerUnit, 81.25) {
		t.Fatalf("eff=%v want 81.25", resp.EffectiveCostPerUnit)
	}
}

func TestEvaluateEndpoint_ValidationError(t *testing.T) {
	srv := NewServer(nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(`{"basePrice":-100}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d want 400", rec.Code)
	}
}

func TestEvaluateEndpoint_BadJSON(t *testing.T) {
	srv := NewServer(nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader("{not json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d want 400", rec.Code)
	}
}

func TestHistoryEndpoint_UnavailableWithoutDB(t *testing.T) {
	srv := NewServer(nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/history", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d want 503", rec.Code)
	}
}
