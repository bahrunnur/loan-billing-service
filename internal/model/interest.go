package model

import (
	"github.com/shopspring/decimal"
)

type PreciseInterestRate struct {
	rate decimal.Decimal
}

func NewPreciseInterestRate(rateStr string) (*PreciseInterestRate, error) {
	rate, err := decimal.NewFromString(rateStr)
	if err != nil {
		return nil, err
	}
	return &PreciseInterestRate{
		rate: rate,
	}, nil
}

func (pir *PreciseInterestRate) CalculateInterest(principal string) (decimal.Decimal, error) {
	precisionPrincipal, err := decimal.NewFromString(principal)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return precisionPrincipal.Mul(pir.rate.Div(decimal.NewFromInt(100))), nil
}
