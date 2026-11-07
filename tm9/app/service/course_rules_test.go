package service

import "testing"

func TestWithinCreditLimit(t *testing.T) {
	cases := []struct {
		taken, adding int
		want          bool
	}{
		{0, 3, true},
		{21, 3, true}, // tepat 24: masih boleh
		{22, 3, false},
		{24, 1, false},
		{24, 0, true},
	}
	for _, c := range cases {
		if got := WithinCreditLimit(c.taken, c.adding); got != c.want {
			t.Errorf("WithinCreditLimit(%d, %d) = %v, want %v", c.taken, c.adding, got, c.want)
		}
	}
}

func TestCreditsAfterSwap(t *testing.T) {
	if got := CreditsAfterSwap(24, 3, 3); got != 24 {
		t.Errorf("tukar kelas 3 SKS dengan 3 SKS pada 24 SKS harus tetap 24, dapat %d", got)
	}
	if got := CreditsAfterSwap(24, 2, 3); got != 25 {
		t.Errorf("harus 25, dapat %d", got)
	}
}
