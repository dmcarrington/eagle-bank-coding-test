package domain

import (
	"encoding/json"
	"math"
)

// Pence stores monetary values as integer pence to avoid float drift.
type Pence int64

func FromGBPFloat(f float64) Pence {
	return Pence(math.Round(f * 100))
}

func (p Pence) GBPFloat() float64 {
	return math.Round(float64(p)) / 100
}

func (p Pence) MarshalJSON() ([]byte, error) {
	return json.Marshal(math.Round(float64(p)/100*100) / 100)
}

func (p *Pence) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*p = FromGBPFloat(f)
	return nil
}
