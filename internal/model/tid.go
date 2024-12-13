package model

import "go.jetify.com/typeid"

type LoanPrefix struct{}

func (LoanPrefix) Prefix() string { return "loan" }

type LoanID struct {
	typeid.TypeID[LoanPrefix]
}
