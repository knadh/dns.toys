package aerial

import (
	"errors"
	"fmt"
	"testing"
)

var tests = []struct {
	l1   Location
	l2   Location
	dist float64
	err  error
}{
	{
		Location{Lat: 30.2458, Long: 75.8421}, // Sangrur
		Location{Lat: 30.2001, Long: 75.6755}, // Longowal
		16.793459061041027,
		nil,
	},
	{
		Location{Lat: 12.9352, Long: 77.6245}, // Kormangala
		Location{Lat: 12.9698, Long: 77.7500}, // Whitefield
		14.132940521067107,
		nil,
	},
	{
		Location{Lat: 12.9716, Long: 77.5946}, // Bengaluru
		Location{Lat: 28.7041, Long: 77.1025}, // New Delhi
		1750.0305628709923,
		nil,
	},
	{
		Location{Lat: -120.9716, Long: 77.5946}, // Wrong Lat
		Location{Lat: 28.7041, Long: 77.1025},   // New Delhi
		0,
		errors.New("-120.9716: lat out of bounds"),
	},
	{
		Location{Lat: 120.9716, Long: 77.5946}, // Wrong Lat
		Location{Lat: 28.7041, Long: 77.1025},  // New Delhi
		0,
		errors.New("120.9716: lat out of bounds"),
	},
	{
		Location{Lat: 12.9716, Long: 277.5946}, // Wrong Long
		Location{Lat: 28.7041, Long: 77.1025},  // New Delhi
		0,
		errors.New("277.5946: long out of bounds"),
	},
	{
		Location{Lat: 12.9716, Long: -277.5946}, // Wrong Long
		Location{Lat: 28.7041, Long: 77.1025},   // New Delhi
		0,
		errors.New("-277.5946: long out of bounds"),
	},
	{
		Location{Lat: 12.9716, Long: 77.5946},  // Bengaluru
		Location{Lat: 128.7041, Long: 77.1025}, // Wrong Lat
		0,
		errors.New("128.7041: lat out of bounds"),
	},
	{
		Location{Lat: 12.9716, Long: 77.5946},   // Bengaluru
		Location{Lat: -128.7041, Long: 77.1025}, // Wrong Lat
		0,
		errors.New("-128.7041: lat out of bounds"),
	},
	{
		Location{Lat: 12.9716, Long: 77.5946},   // Bengaluru
		Location{Lat: 28.7041, Long: -187.1025}, // Wrong Long
		0,
		errors.New("-187.1025: long out of bounds"),
	},
	{
		Location{Lat: 12.9716, Long: 77.5946},  // Bengaluru
		Location{Lat: 28.7041, Long: 187.1025}, // Wrong Long
		0,
		errors.New("187.1025: long out of bounds"),
	},
	{
		Location{Lat: -120.9716, Long: 77.5946}, // Wrong Lat
		Location{Lat: 28.7041, Long: 187.1025},  // Wrong Long
		0,
		errors.New("-120.9716: lat out of bounds"),
	},
	{
		Location{Lat: -120.9716, Long: 277.5946}, // Wrong Lat Long
		Location{Lat: 28.7041, Long: 187.1025},   // Wrong Long
		0,
		errors.New("-120.9716: lat out of bounds 277.5946: long out of bounds"),
	},
}

func TestCalculateAerialDistance(t *testing.T) {
	for n, input := range tests {
		d, err := Calculate(input.l1, input.l2)

		if err != nil && input.err == nil {
			t.Errorf("fail %d: %v %v -> didn't get err", n, input.l1, input.l2)
		}

		if err != nil && input.err != nil && input.err.Error() != err.Error() {
			t.Errorf("fail %d: %v %v ->want error:%s->got error:%s", n, input.l1, input.l2, input.err.Error(), err.Error())
		}

		if fmt.Sprintf("%0.5f", input.dist) != fmt.Sprintf("%0.5f", d) {
			t.Errorf("fail %d: want %v %v -> %v got %v", n, input.l1, input.l2, input.dist, d)
		}
	}
}
