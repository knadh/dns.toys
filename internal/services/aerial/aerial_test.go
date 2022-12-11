package aerial_test

import (
	"errors"
	"testing"

	"github.com/knadh/dns.toys/internal/services/aerial"
)

var tests = []struct {
	l1          aerial.Location
	l2          aerial.Location
	d           float64
	e           error
}{
	{
		aerial.Location{Lat: 30.2458, Long: 75.8421}, // Sangrur
		aerial.Location{Lat: 30.2001, Long: 75.6755}, // Longowal
		16.793459061041027,
		nil,
	},
	{
		aerial.Location{Lat: 12.9352, Long: 77.6245}, // Kormangala
		aerial.Location{Lat: 12.9698, Long: 77.7500}, // Whitefield
		14.132940521067107,
		nil,
	},
	{
		aerial.Location{Lat: 12.9716, Long: 77.5946}, // Bengaluru
		aerial.Location{Lat: 28.7041, Long: 77.1025}, // New Delhi
		1750.0305628709923,
		nil,
	},
	{
		aerial.Location{Lat: -120.9716, Long: 77.5946}, // Wrong Lat
		aerial.Location{Lat: 28.7041, Long: 77.1025},   // New Delhi
		0,
		errors.New("-120.9716 lat out of bounds"),
	},
	{
		aerial.Location{Lat: 120.9716, Long: 77.5946}, // Wrong Lat
		aerial.Location{Lat: 28.7041, Long: 77.1025},  // New Delhi
		0,
		errors.New("120.9716 lat out of bounds"),
	},
	{
		aerial.Location{Lat: 12.9716, Long: 277.5946}, // Wrong Lng
		aerial.Location{Lat: 28.7041, Long: 77.1025},  // New Delhi
		0,
		errors.New("277.5946 lng out of bounds"),
	},
	{
		aerial.Location{Lat: 12.9716, Long: -277.5946}, // Wrong Lng
		aerial.Location{Lat: 28.7041, Long: 77.1025},   // New Delhi
		0,
		errors.New("-277.5946 lng out of bounds"),
	},
	{
		aerial.Location{Lat: 12.9716, Long: 77.5946},  // Bengaluru
		aerial.Location{Lat: 128.7041, Long: 77.1025}, // Wrong Lat
		0,
		errors.New("128.7041 lat out of bounds"),
	},
	{
		aerial.Location{Lat: 12.9716, Long: 77.5946},   // Bengaluru
		aerial.Location{Lat: -128.7041, Long: 77.1025}, // Wrong Lat
		0,
		errors.New("-128.7041 lat out of bounds"),
	},
	{
		aerial.Location{Lat: 12.9716, Long: 77.5946},   // Bengaluru
		aerial.Location{Lat: 28.7041, Long: -187.1025}, // Wrong Lng
		0,
		errors.New("-187.1025 lng out of bounds"),
	},
	{
		aerial.Location{Lat: 12.9716, Long: 77.5946},  // Bengaluru
		aerial.Location{Lat: 28.7041, Long: 187.1025}, // Wrong Lng
		0,
		errors.New("187.1025 lng out of bounds"),
	},
	{
		aerial.Location{Lat: -120.9716, Long: 77.5946},  // Wrong Lat
		aerial.Location{Lat: 28.7041, Long: 187.1025}, // Wrong Lng
		0,
		errors.New("-120.9716 lat out of bounds; 187.1025 lng out of bounds"),
	},
	{
		aerial.Location{Lat: -120.9716, Long: 277.5946},  // Wrong Lat Lng
		aerial.Location{Lat: 28.7041, Long: 187.1025}, // Wrong Lng
		0,
		errors.New("-120.9716 lat out of bounds 277.5946 lng out of bounds; 187.1025 lng out of bounds"),
	},
}

func TestCalculateAerialDistance(t *testing.T) {
	for _, input := range tests {
		d, e := aerial.CalculateAerialDistance(input.l1, input.l2)

		if e != nil && input.e == nil {
			t.Errorf("fail: %v %v -> expected %v -> got error:%s;", input.l1, input.l2, input.d, e.Error())
		}

		if e != nil && input.e != nil && input.e.Error() != e.Error() {
			t.Errorf("fail: %v %v ->expected error:%s-> got error:%s;", input.l1, input.l2, input.e.Error(), e.Error())
		}

		if input.d != d {
			t.Errorf("fail: want %v %v -> %v got %v", input.l1, input.l2, input.d, d)
		}
	}
}
