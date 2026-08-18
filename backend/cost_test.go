package backend

import (
	"math"
	"testing"
)

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestEffectiveCostPerUnit_FullScheme(t *testing.T) {
	// base 100, trade 10%, cash 5%, buy 10 get 2 free, freight 120 extra.
	// 100*.9=90 ; 90*.95=85.5 ; 85.5*10/12=71.25 ; 120/12=10 ; total 81.25
	eff, bd := EffectiveCostPerUnit(Input{
		BasePrice: 100, TradeDiscountPct: 10, CashDiscountPct: 5,
		FreeBuyQty: 10, FreeGetQty: 2, Freight: 120, FreightExtra: true,
	})
	if !almost(bd.AfterTrade, 90) || !almost(bd.AfterCash, 85.5) ||
		!almost(bd.AfterFreeGoods, 71.25) || !almost(bd.FreightPerUnit, 10) || !almost(eff, 81.25) {
		t.Fatalf("got %+v eff=%v, want effective 81.25", bd, eff)
	}
}

func TestEffectiveCostPerUnit_FreightIncluded(t *testing.T) {
	eff, bd := EffectiveCostPerUnit(Input{BasePrice: 100, FreeBuyQty: 1, Freight: 50, FreightExtra: false})
	if !almost(bd.FreightPerUnit, 0) || !almost(eff, 100) {
		t.Fatalf("freight-included should not add freight; eff=%v", eff)
	}
}

func TestEffectiveCostPerUnit_ZeroFreeGoodsFactorOne(t *testing.T) {
	eff, _ := EffectiveCostPerUnit(Input{BasePrice: 100, FreeBuyQty: 0, FreeGetQty: 5})
	if !almost(eff, 100) {
		t.Fatalf("X=0 => free-goods factor 1; eff=%v want 100", eff)
	}
}

func TestValidate(t *testing.T) {
	if err := (Input{BasePrice: 100, FreeBuyQty: 1}).Validate(); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	bad := []Input{
		{BasePrice: -1}, {TradeDiscountPct: 101}, {CashDiscountPct: -5}, {Freight: -1}, {FreeBuyQty: -2},
	}
	for i, in := range bad {
		if err := in.Validate(); err == nil {
			t.Fatalf("bad input %d accepted", i)
		}
	}
}
