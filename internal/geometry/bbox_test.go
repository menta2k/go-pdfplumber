package geometry

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestIntersect(t *testing.T) {
	tests := []struct {
		name   string
		a, b   model.BBox
		want   model.BBox
		wantOK bool
	}{
		{
			name:   "overlapping boxes",
			a:      model.BBox{X0: 0, Y0: 0, X1: 10, Y1: 10},
			b:      model.BBox{X0: 5, Y0: 5, X1: 15, Y1: 15},
			want:   model.BBox{X0: 5, Y0: 5, X1: 10, Y1: 10},
			wantOK: true,
		},
		{
			name:   "no overlap",
			a:      model.BBox{X0: 0, Y0: 0, X1: 5, Y1: 5},
			b:      model.BBox{X0: 10, Y0: 10, X1: 15, Y1: 15},
			want:   model.BBox{},
			wantOK: false,
		},
		{
			name:   "contained",
			a:      model.BBox{X0: 0, Y0: 0, X1: 20, Y1: 20},
			b:      model.BBox{X0: 5, Y0: 5, X1: 10, Y1: 10},
			want:   model.BBox{X0: 5, Y0: 5, X1: 10, Y1: 10},
			wantOK: true,
		},
		{
			name:   "touching edge",
			a:      model.BBox{X0: 0, Y0: 0, X1: 5, Y1: 5},
			b:      model.BBox{X0: 5, Y0: 0, X1: 10, Y1: 5},
			want:   model.BBox{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Intersect(tt.a, tt.b)
			if ok != tt.wantOK {
				t.Fatalf("Intersect ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("Intersect = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	a := model.BBox{X0: 0, Y0: 0, X1: 5, Y1: 5}
	b := model.BBox{X0: 3, Y0: 3, X1: 10, Y1: 10}
	want := model.BBox{X0: 0, Y0: 0, X1: 10, Y1: 10}

	got := Union(a, b)
	if got != want {
		t.Errorf("Union = %v, want %v", got, want)
	}
}

func TestUnionAll(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := UnionAll(nil)
		if got != (model.BBox{}) {
			t.Errorf("UnionAll(nil) = %v, want zero", got)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		boxes := []model.BBox{
			{X0: 0, Y0: 0, X1: 5, Y1: 5},
			{X0: 10, Y0: 10, X1: 15, Y1: 15},
			{X0: 3, Y0: 3, X1: 12, Y1: 12},
		}
		want := model.BBox{X0: 0, Y0: 0, X1: 15, Y1: 15}
		got := UnionAll(boxes)
		if got != want {
			t.Errorf("UnionAll = %v, want %v", got, want)
		}
	})
}

func TestNormalize(t *testing.T) {
	b := model.BBox{X0: 10, Y0: 10, X1: 0, Y1: 0}
	got := Normalize(b)
	want := model.BBox{X0: 0, Y0: 0, X1: 10, Y1: 10}
	if got != want {
		t.Errorf("Normalize = %v, want %v", got, want)
	}
}

func TestExpand(t *testing.T) {
	b := model.BBox{X0: 5, Y0: 5, X1: 10, Y1: 10}
	got := Expand(b, 2)
	want := model.BBox{X0: 3, Y0: 3, X1: 12, Y1: 12}
	if got != want {
		t.Errorf("Expand = %v, want %v", got, want)
	}
}
