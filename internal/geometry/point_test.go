package geometry

import (
	"math"
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name string
		a, b model.Point
		want float64
	}{
		{"same point", model.Point{X: 0, Y: 0}, model.Point{X: 0, Y: 0}, 0},
		{"horizontal", model.Point{X: 0, Y: 0}, model.Point{X: 3, Y: 0}, 3},
		{"vertical", model.Point{X: 0, Y: 0}, model.Point{X: 0, Y: 4}, 4},
		{"3-4-5", model.Point{X: 0, Y: 0}, model.Point{X: 3, Y: 4}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-10 {
				t.Errorf("Distance = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSegmentIntersection(t *testing.T) {
	tests := []struct {
		name                                    string
		ax0, ay0, ax1, ay1, bx0, by0, bx1, by1 float64
		wantPt                                  model.Point
		wantOK                                  bool
	}{
		{
			name: "cross at origin",
			ax0: -1, ay0: 0, ax1: 1, ay1: 0,
			bx0: 0, by0: -1, bx1: 0, by1: 1,
			wantPt: model.Point{X: 0, Y: 0},
			wantOK: true,
		},
		{
			name: "parallel horizontal",
			ax0: 0, ay0: 0, ax1: 10, ay1: 0,
			bx0: 0, by0: 5, bx1: 10, by1: 5,
			wantOK: false,
		},
		{
			name: "non-intersecting",
			ax0: 0, ay0: 0, ax1: 1, ay1: 0,
			bx0: 2, by0: -1, bx1: 2, by1: 1,
			wantOK: false,
		},
		{
			name: "T-intersection at endpoint",
			ax0: 0, ay0: 5, ax1: 10, ay1: 5,
			bx0: 5, by0: 0, bx1: 5, by1: 5,
			wantPt: model.Point{X: 5, Y: 5},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SegmentIntersection(tt.ax0, tt.ay0, tt.ax1, tt.ay1, tt.bx0, tt.by0, tt.bx1, tt.by1)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if math.Abs(got.X-tt.wantPt.X) > 1e-6 || math.Abs(got.Y-tt.wantPt.Y) > 1e-6 {
					t.Errorf("point = %v, want %v", got, tt.wantPt)
				}
			}
		})
	}
}

func TestNearlyEqual(t *testing.T) {
	if !NearlyEqual(1.0, 1.0001, 0.001) {
		t.Error("expected 1.0 ≈ 1.0001 with tolerance 0.001")
	}
	if NearlyEqual(1.0, 2.0, 0.5) {
		t.Error("expected 1.0 ≠ 2.0 with tolerance 0.5")
	}
}
