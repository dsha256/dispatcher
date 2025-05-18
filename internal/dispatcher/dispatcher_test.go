package dispatcher_test

import (
	"reflect"
	"testing"

	"github.com/dsha256/dispatcher/internal/dispatcher"
)

func TestReconstructItinerary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		name     string
		tickets  [][]string
		expected []string
	}{
		{
			name:     "Standard itinerary",
			tickets:  [][]string{{"LAX", "DXB"}, {"JFK", "LAX"}, {"SFO", "SJC"}, {"DXB", "SFO"}},
			expected: []string{"JFK", "LAX", "DXB", "SFO", "SJC"},
			err:      nil,
		},
		{
			name:     "Multiple possible paths",
			tickets:  [][]string{{"JFK", "SFO"}, {"JFK", "ATL"}, {"SFO", "ATL"}, {"ATL", "JFK"}},
			expected: []string{"JFK", "ATL", "JFK", "SFO", "ATL"},
			err:      nil,
		},
		{
			name:     "Single ticket",
			tickets:  [][]string{{"SFO", "JFK"}},
			expected: []string{"SFO", "JFK"},
			err:      nil,
		},
		{
			name:     "Empty input",
			tickets:  [][]string{},
			expected: []string{},
			err:      nil,
		},
		{
			name:     "Different starting point",
			tickets:  [][]string{{"SFO", "LAX"}, {"LAX", "JFK"}, {"JFK", "SFO"}},
			expected: []string{"JFK", "SFO", "LAX", "JFK"},
			err:      dispatcher.ErrDifferentStartingPoints,
		},
		{
			name:     "Cycle in itinerary",
			tickets:  [][]string{{"JFK", "SFO"}, {"SFO", "LAX"}, {"LAX", "JFK"}, {"JFK", "ATL"}},
			expected: []string{"JFK", "SFO", "LAX", "JFK", "ATL"},
			err:      nil,
		},
		{
			name:     "Multiple same destination",
			tickets:  [][]string{{"JFK", "SFO"}, {"JFK", "ATL"}, {"JFK", "SFO"}, {"SFO", "LAX"}, {"ATL", "LAX"}},
			expected: []string{"JFK", "ATL", "LAX", "JFK", "SFO", "SFO"},
			err:      dispatcher.ErrMultipleSameDestination,
		},
		{
			name:     "Duplicate tickets turned into cycle",
			tickets:  [][]string{{"JFK", "SFO"}, {"JFK", "SFO"}, {"SFO", "LAX"}, {"LAX", "ATL"}},
			expected: nil,
			err:      dispatcher.ErrMultipleSameDestination,
		},
		{
			name:     "Longer complex itinerary",
			tickets:  [][]string{{"A", "B"}, {"B", "C"}, {"C", "D"}, {"D", "E"}, {"E", "F"}, {"F", "A"}, {"A", "G"}},
			expected: []string{"A", "B", "C", "D", "E", "F", "A", "G"},
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := dispatcher.ReconstructItinerary(tt.tickets)
			if tt.err != nil && err == nil {
				t.Errorf("reconstructItinerary(%v) = %v; want %v", tt.tickets, err, tt.err)
			}

			if tt.err == nil && err != nil {
				t.Errorf("reconstructItinerary(%v) = %v; want %v", tt.tickets, err, tt.err)
			}

			if err != nil && tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("reconstructItinerary(%v) = %v; want %v", tt.tickets, err, tt.err)
				}
			}

			if tt.err == nil {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("reconstructItinerary(%v) = %v; want %v", tt.tickets, result, tt.expected)
				}
			}
		})
	}
}
