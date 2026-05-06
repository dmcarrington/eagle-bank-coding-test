package domain

import (
	"encoding/json"
	"testing"
)

func TestFromGBPFloat(t *testing.T) {
	cases := []struct {
		input float64
		want  Pence
	}{
		{0.00, 0},
		{1.00, 100},
		{1.99, 199},
		{10.005, 1001}, // rounds to nearest penny: 1000.5 → 1001
		{100.00, 10000},
		{10000.00, 1000000},
	}
	for _, tc := range cases {
		got := FromGBPFloat(tc.input)
		if got != tc.want {
			t.Errorf("FromGBPFloat(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestPenceGBPFloat(t *testing.T) {
	if got := Pence(100).GBPFloat(); got != 1.0 {
		t.Errorf("Pence(100).GBPFloat() = %v, want 1.0", got)
	}
	if got := Pence(199).GBPFloat(); got != 1.99 {
		t.Errorf("Pence(199).GBPFloat() = %v, want 1.99", got)
	}
}

func TestPenceMarshalJSON(t *testing.T) {
	type wrapper struct {
		Amount Pence `json:"amount"`
	}
	w := wrapper{Amount: Pence(1999)}
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"amount":19.99}` {
		t.Errorf("got %s, want {\"amount\":19.99}", string(b))
	}
}

func TestPenceUnmarshalJSON(t *testing.T) {
	type wrapper struct {
		Amount Pence `json:"amount"`
	}
	var w wrapper
	if err := json.Unmarshal([]byte(`{"amount":19.99}`), &w); err != nil {
		t.Fatal(err)
	}
	if w.Amount != Pence(1999) {
		t.Errorf("got %v, want 1999", w.Amount)
	}
}
