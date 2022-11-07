package aerial_test

import (
	"testing"
	"github.com/knadh/dns.toys/internal/services/aerial"
)

var tests = []struct {
	l1 aerial.Location
	l2 aerial.Location
	d float64
}{
	{
		aerial.Location{Lat: 30.2458, Lng: 75.8421}, // Sangrur
		aerial.Location{Lat: 30.2001, Lng: 75.6755}, // Longowal
		16.793459061041027,
	},
	{
		aerial.Location{Lat: 12.9352, Lng: 77.6245},  // Kormangala
		aerial.Location{Lat: 12.9698, Lng: 77.7500}, // Whitefield
		14.132940521067107,
	},
	{
		aerial.Location{Lat: 12.9716, Lng: 77.5946}, // Bengaluru
		aerial.Location{Lat: 28.7041, Lng: 77.1025}, // New Delhi
		1750.0305628709923,
	},
}

func TestCalculateAerialDistance(t *testing.T) {
	for _, input := range tests {
		d, err := aerial.CalculateAerialDistance(input.l1, input.l2)
		if err != nil {
			t.Errorf("fail: %v %v -> %v got error: %v", input.l1, input.l2, input.d, err)
		}

		if input.d != d {
			t.Errorf("fail: want %v %v -> %v got %v", input.l1, input.l2, input.d, d)
		}
	}
}