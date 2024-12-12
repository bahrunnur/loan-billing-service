package currency_test

import (
	"testing"

	"github.com/bahrunnur/loan-billing-service/pkg/currency"
)

func TestNewRupiah(t *testing.T) {
	testCases := []struct {
		name     string
		rupiah   int
		sen      int
		expected currency.Rupiah
	}{
		{
			name:     "Basic creation",
			rupiah:   10000,
			sen:      50,
			expected: currency.Rupiah(1000050),
		},
		{
			name:     "Zero values",
			rupiah:   0,
			sen:      0,
			expected: currency.Rupiah(0),
		},
		{
			name:     "Large value",
			rupiah:   1000000,
			sen:      99,
			expected: currency.Rupiah(100000099),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := currency.NewRupiah(tc.rupiah, tc.sen)
			if result != tc.expected {
				t.Errorf("NewRupiah(%d, %d) = %d, want %d",
					tc.rupiah, tc.sen, result, tc.expected)
			}
		})
	}
}

func TestRupiah(t *testing.T) {
	testCases := []struct {
		name     string
		value    currency.Rupiah
		expected int
	}{
		{
			name:     "Basic rupiah extraction",
			value:    currency.Rupiah(1000050),
			expected: 10000,
		},
		{
			name:     "Zero value",
			value:    currency.Rupiah(0),
			expected: 0,
		},
		{
			name:     "Large value",
			value:    currency.Rupiah(100000099),
			expected: 1000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.Rupiah()
			if result != tc.expected {
				t.Errorf("Rupiah() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestSen(t *testing.T) {
	testCases := []struct {
		name     string
		value    currency.Rupiah
		expected int
	}{
		{
			name:     "Basic sen extraction",
			value:    currency.Rupiah(1000050),
			expected: 50,
		},
		{
			name:     "Zero value",
			value:    currency.Rupiah(0),
			expected: 0,
		},
		{
			name:     "Large value",
			value:    currency.Rupiah(100000099),
			expected: 99,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.Sen()
			if result != tc.expected {
				t.Errorf("Sen() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	testCases := []struct {
		name     string
		a        currency.Rupiah
		b        currency.Rupiah
		expected currency.Rupiah
	}{
		{
			name:     "Basic addition",
			a:        currency.Rupiah(1000050),
			b:        currency.Rupiah(2000075),
			expected: currency.Rupiah(3000125),
		},
		{
			name:     "Zero addition",
			a:        currency.Rupiah(1000000),
			b:        currency.Rupiah(0),
			expected: currency.Rupiah(1000000),
		},
		{
			name:     "Large values",
			a:        currency.Rupiah(1000000000),
			b:        currency.Rupiah(5000000099),
			expected: currency.Rupiah(6000000099),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.a.Add(tc.b)
			if result != tc.expected {
				t.Errorf("Add(%v) = %v, want %v", tc.b, result, tc.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	testCases := []struct {
		name     string
		a        currency.Rupiah
		b        currency.Rupiah
		expected currency.Rupiah
	}{
		{
			name:     "Basic subtraction",
			a:        currency.Rupiah(2075),
			b:        currency.Rupiah(1050),
			expected: currency.Rupiah(1025),
		},
		{
			name:     "Zero subtraction",
			a:        currency.Rupiah(1000),
			b:        currency.Rupiah(0),
			expected: currency.Rupiah(1000),
		},
		{
			name:     "Large values",
			a:        currency.Rupiah(100099),
			b:        currency.Rupiah(50050),
			expected: currency.Rupiah(50049),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.a.Subtract(tc.b)
			if result != tc.expected {
				t.Errorf("Subtract(%v) = %v, want %v", tc.b, result, tc.expected)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	testCases := []struct {
		name     string
		value    currency.Rupiah
		factor   int
		expected currency.Rupiah
	}{
		{
			name:     "Basic multiplication",
			value:    currency.Rupiah(1000050),
			factor:   2,
			expected: currency.Rupiah(2000100),
		},
		{
			name:     "Zero multiplication",
			value:    currency.Rupiah(100000),
			factor:   0,
			expected: currency.Rupiah(0),
		},
		{
			name:     "Large factor",
			value:    currency.Rupiah(10000000050),
			factor:   10,
			expected: currency.Rupiah(100000000500),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.Multiply(tc.factor)
			if result != tc.expected {
				t.Errorf("Multiply(%d) = %v, want %v", tc.factor, result, tc.expected)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	testCases := []struct {
		name     string
		value    currency.Rupiah
		divisor  int
		expected currency.Rupiah
	}{
		{
			name:     "Basic division",
			value:    currency.Rupiah(1000050),
			divisor:  2,
			expected: currency.Rupiah(500025),
		},
		{
			name:     "Large values",
			value:    currency.Rupiah(100000050),
			divisor:  3,
			expected: currency.Rupiah(33333350),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.Divide(tc.divisor)
			if result != tc.expected {
				t.Errorf("Divide(%d) = %v, want %v", tc.divisor, result, tc.expected)
			}
		})
	}
}

func TestString(t *testing.T) {
	testCases := []struct {
		name     string
		value    currency.Rupiah
		expected string
	}{
		{
			name:     "Basic representation",
			value:    currency.Rupiah(1000050),
			expected: "Rp 10000,50",
		},
		{
			name:     "Zero value",
			value:    currency.Rupiah(0),
			expected: "Rp 0,00",
		},
		{
			name:     "Large value",
			value:    currency.Rupiah(100000099),
			expected: "Rp 1000000,99",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.String()
			if result != tc.expected {
				t.Errorf("String() = %s, want %s", result, tc.expected)
			}
		})
	}
}
