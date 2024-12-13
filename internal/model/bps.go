package model

// BPS is a basis point
type BPS int

func (b BPS) ToPercentage() int { // TODO: use a better way like math/big or Decimal
	return int(b) / 100
}

func FromPercentage(p int) BPS {
	return BPS(p * 100)
}
