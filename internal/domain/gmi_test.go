package domain

import (
	"math"
	"testing"
)

func TestCalculateGMI(t *testing.T) {
	tests := []struct {
		name        string
		averageMgDl float64
		wantNil     bool
		wantValue   float64
	}{
		{"typical value 154 mg/dL", 154, false, 6.99},
		{"low value 100 mg/dL", 100, false, 5.70},
		{"high value 250 mg/dL", 250, false, 9.29},
		{"zero returns nil", 0, true, 0},
		{"negative returns nil", -10, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateGMI(tt.averageMgDl)
			if tt.wantNil {
				if got != nil {
					t.Errorf("CalculateGMI(%v) = %v, want nil", tt.averageMgDl, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("CalculateGMI(%v) = nil, want %v", tt.averageMgDl, tt.wantValue)
			}
			if math.Abs(*got-tt.wantValue) > 0.01 {
				t.Errorf("CalculateGMI(%v) = %v, want ~%v", tt.averageMgDl, *got, tt.wantValue)
			}
		})
	}
}
