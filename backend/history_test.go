package backend

import (
	"context"
	"os"
	"testing"
)

// Integration test for the Postgres history store. Runs only when DATABASE_URL
// is set (scripts/test.sh spins up a throwaway Postgres, then tears it down).
func TestPostgresHistory_RoundTrip(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}
	store, err := NewPostgresStore(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer store.Close()

	in := Input{BasePrice: 100, TradeDiscountPct: 10, FreeBuyQty: 1}
	eff, _ := EffectiveCostPerUnit(in)
	saved, err := store.Save(Evaluation{Input: in, Effective: eff, Note: "test"})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.ID == 0 {
		t.Fatal("expected assigned id")
	}
	items, err := store.List(10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) == 0 || items[0].ID != saved.ID {
		t.Fatalf("expected saved row first, got %+v", items)
	}
	if !almost(items[0].Effective, eff) {
		t.Fatalf("effective roundtrip mismatch: %v vs %v", items[0].Effective, eff)
	}
}
