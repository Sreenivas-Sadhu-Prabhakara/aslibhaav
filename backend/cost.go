package backend

import "fmt"

// Input holds the terms of a distributor scheme for a single item.
type Input struct {
	BasePrice        float64 `json:"basePrice"`
	TradeDiscountPct float64 `json:"tradeDiscountPct"`
	CashDiscountPct  float64 `json:"cashDiscountPct"`
	FreeBuyQty       int     `json:"freeBuyQty"`
	FreeGetQty       int     `json:"freeGetQty"`
	Freight          float64 `json:"freight"`
	FreightExtra     bool    `json:"freightExtra"`
}

// Breakdown captures each intermediate value on the way to the effective cost,
// so the shopkeeper can see exactly how the "real price" was derived.
type Breakdown struct {
	AfterTrade     float64 `json:"afterTrade"`
	AfterCash      float64 `json:"afterCash"`
	AfterFreeGoods float64 `json:"afterFreeGoods"`
	FreightPerUnit float64 `json:"freightPerUnit"`
	Effective      float64 `json:"effective"`
}

// Validate reports whether the Input is well formed.
func (in Input) Validate() error {
	if in.BasePrice < 0 {
		return fmt.Errorf("base price cannot be negative")
	}
	if in.Freight < 0 {
		return fmt.Errorf("freight cannot be negative")
	}
	if in.FreeBuyQty < 0 || in.FreeGetQty < 0 {
		return fmt.Errorf("free-goods quantities cannot be negative")
	}
	if in.TradeDiscountPct < 0 || in.TradeDiscountPct > 100 {
		return fmt.Errorf("trade discount must be between 0 and 100")
	}
	if in.CashDiscountPct < 0 || in.CashDiscountPct > 100 {
		return fmt.Errorf("cash discount must be between 0 and 100")
	}
	return nil
}

// EffectiveCostPerUnit computes the true effective cost per unit of a scheme.
// Trade and cash discounts compound sequentially; buy-X-get-Y free goods dilute
// the cost over X+Y received units; freight is allocated per received unit only
// when it is charged on top (FreightExtra).
func EffectiveCostPerUnit(in Input) (float64, Breakdown) {
	afterTrade := in.BasePrice * (1 - in.TradeDiscountPct/100)
	afterCash := afterTrade * (1 - in.CashDiscountPct/100)

	freeGoodsFactor := 1.0
	unitsReceived := in.FreeBuyQty + in.FreeGetQty
	if in.FreeBuyQty > 0 && unitsReceived > 0 {
		freeGoodsFactor = float64(in.FreeBuyQty) / float64(unitsReceived)
	}
	afterFreeGoods := afterCash * freeGoodsFactor

	freightPerUnit := 0.0
	if in.FreightExtra {
		denom := unitsReceived
		if denom <= 0 {
			denom = 1
		}
		freightPerUnit = in.Freight / float64(denom)
	}

	effective := afterFreeGoods + freightPerUnit
	return effective, Breakdown{
		AfterTrade:     afterTrade,
		AfterCash:      afterCash,
		AfterFreeGoods: afterFreeGoods,
		FreightPerUnit: freightPerUnit,
		Effective:      effective,
	}
}
