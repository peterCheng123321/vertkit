package domain

import "testing"

func TestINRUsesTwoMinorUnits(t *testing.T) {
	money, err := NewMoney(1234, CurrencyINR)
	if err != nil {
		t.Fatalf("new money: %v", err)
	}
	if got, want := MinorUnitsPerUnit(CurrencyINR), int64(100); got != want {
		t.Fatalf("minor units = %d, want %d", got, want)
	}
	if got, want := money.Format(), "INR 12.34"; got != want {
		t.Fatalf("format = %q, want %q", got, want)
	}
}
