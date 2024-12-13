package model_test

import (
	"testing"

	"github.com/bahrunnur/loan-billing-service/internal/model"
	"github.com/bahrunnur/loan-billing-service/pkg/currency"
	"github.com/shopspring/decimal"
)

func TestInterest(t *testing.T) {
	// Decimal approach (recommended) but do it later
	rate, _ := model.NewPreciseInterestRate("10.00")
	rp := currency.NewRupiah(10000, 0)
	interest, _ := rate.CalculateInterest(rp.DecimalString())

	expected := decimal.NewFromInt(1000)
	if interest.String() != expected.String() {
		t.Errorf("CalculateInterest() = %d, want %d", interest, expected)
	}
}
