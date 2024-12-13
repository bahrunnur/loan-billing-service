package currency

import "fmt"

// source: https://jdih.kominfo.go.id/produk_hukum/view/id/92/t/undangundang+nomor+7+tahun+2011+tanggal+28+juni+2011
// mirror: http://www.flevin.com/id/lgso/legislation/Mirror/czozMToiZD0yMDAwKzExJmY9dXU3LTIwMTFidC5odG0manM9MSI7.html

const FRACTION = 100 // (Pasal 3 ayat 2: Satu Rupiah adalah 100 (seratus) sen)

// Rupiah represents a monetary value as an integer with 'sen' as the fraction part
type Rupiah int

// NewRupiah creates a new Money value for Rupiah
func NewRupiah(rupiah int, sen int) Rupiah {
	return Rupiah(rupiah*FRACTION + sen)
}

// Rupiah return the amount without sen
func (m Rupiah) Rupiah() int {
	return int(m) / FRACTION
}

// Sen return the sen part, although at the moment we don't use that in real life, it can still be represented in banking
func (m Rupiah) Sen() int {
	return int(m) % FRACTION
}

// Add combines two monetary values
func (m Rupiah) Add(other Rupiah) Rupiah {
	return m + other
}

// Subtract reduces one monetary value from another
func (m Rupiah) Subtract(other Rupiah) Rupiah {
	return m - other
}

// Multiply scales a monetary value
func (m Rupiah) Multiply(factor int) Rupiah {
	return Rupiah(int(m) * factor)
}

// Divide performs division with integer arithmetic
func (m Rupiah) Divide(divisor int) Rupiah {
	if divisor == 0 {
		return m
	}

	return Rupiah(int(m) / divisor)
}

// DecimalString representation for decimal package
func (m Rupiah) DecimalString() string {
	return fmt.Sprintf("%d.%02d", m.Rupiah(), m.Sen())
}

// String representation for printing
func (m Rupiah) String() string {
	return fmt.Sprintf("Rp %d,%02d", m.Rupiah(), m.Sen())
}
