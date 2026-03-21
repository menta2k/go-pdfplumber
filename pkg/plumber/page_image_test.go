package plumber

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestNewPageImage(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	img := pi.ToImage()

	if img.Bounds().Dx() != 612 || img.Bounds().Dy() != 792 {
		t.Errorf("image size = %dx%d, want 612x792", img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Background should be white
	r, g, b, _ := img.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("background color = (%d,%d,%d), want white", r>>8, g>>8, b>>8)
	}
}

func TestNewPageImageResolution(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page, PageImageOptions{Resolution: 144}) // 2x
	img := pi.ToImage()

	// At 144 DPI, a 612x792pt page = 1224x1584px
	if img.Bounds().Dx() != 1224 || img.Bounds().Dy() != 1584 {
		t.Errorf("2x image size = %dx%d, want 1224x1584", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestDrawRect(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	pi2 := pi.DrawRect(BBox{X0: 100, Y0: 100, X1: 200, Y1: 200}, DrawStyle{
		StrokeColor: color.RGBA{R: 255, A: 255},
		StrokeWidth: 2,
	})

	// Original should be unchanged (immutability)
	origImg := pi.ToImage()
	newImg := pi2.ToImage()

	// The drawn rect at PDF (100,100)-(200,200) maps to pixel (100,592)-(200,692)
	// Check a pixel on the rect border in the new image has red
	pr, _, _, _ := newImg.At(100, 592).RGBA()
	if pr>>8 != 255 {
		t.Error("new image should have red pixel on rect border")
	}

	// Original should still be white at that location
	or, og, ob, _ := origImg.At(100, 592).RGBA()
	if or>>8 != 255 || og>>8 != 255 || ob>>8 != 255 {
		t.Error("original image should still be white")
	}
}

func TestDrawLine(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	pi2 := pi.DrawLine(0, 396, 612, 396, DrawStyle{
		StrokeColor: color.RGBA{R: 255, A: 255},
	})

	img := pi2.ToImage()
	// Midpoint of page (y=396 in PDF = y=396 in pixels at 72dpi)
	r, _, _, _ := img.At(300, 396).RGBA()
	if r>>8 != 255 {
		t.Error("expected red pixel on drawn line")
	}
}

func TestDrawCircle(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	pi2 := pi.DrawCircle(Point{X: 306, Y: 396}, 20, DrawStyle{
		StrokeColor: color.RGBA{G: 255, A: 255},
		FillColor:   color.RGBA{G: 200, A: 100},
	})

	img := pi2.ToImage()
	_ = img
}

func TestDrawChars(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	pi2 := pi.DrawChars(page.Chars(), StyleCharBBox)

	img := pi2.ToImage()
	// Should have drawn something around char positions
	if img.Bounds().Dx() != 612 {
		t.Error("unexpected image width")
	}
}

func TestDrawWords(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	words := page.ExtractWords()
	pi := NewPageImage(page)
	pi2 := pi.DrawWords(words, StyleWordBBox)

	img := pi2.ToImage()
	_ = img
}

func TestDrawTable(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	tables := page.FindTables()
	if len(tables) == 0 {
		t.Fatal("no tables found")
	}

	pi := NewPageImage(page)
	pi2 := pi.DrawPageRects(DrawStyle{
		StrokeColor: color.RGBA{B: 200, A: 200},
		StrokeWidth: 1,
	})
	pi3 := pi2.DrawTable(tables[0], StyleTableCell)

	img := pi3.ToImage()
	_ = img
}

func TestSavePNG(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	pi = pi.DrawChars(page.Chars(), StyleCharBBox)
	pi = pi.DrawWords(page.ExtractWords(), StyleWordBBox)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_output.png")
	err = pi.Save(path)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("saved PNG is empty")
	}
	t.Logf("Saved PNG: %s (%d bytes)", path, info.Size())
}

func TestSaveTableDebugPNG(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	tables := page.FindTables()
	if len(tables) == 0 {
		t.Fatal("no tables found")
	}

	pi := NewPageImage(page)
	pi = pi.DrawPageRects(DrawStyle{StrokeColor: color.RGBA{B: 150, A: 200}, StrokeWidth: 1})
	pi = pi.DrawTable(tables[0], StyleTableCell)
	pi = pi.DrawWords(page.ExtractWords(), StyleWordBBox)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "table_debug.png")
	err = pi.Save(path)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	t.Logf("Table debug PNG: %s (%d bytes)", path, info.Size())
}

func TestDrawImmutability(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	pi := NewPageImage(page)
	original := pi.ToImage()

	// Draw on a copy
	_ = pi.DrawRect(BBox{X0: 0, Y0: 0, X1: 612, Y1: 792}, DrawStyle{
		StrokeColor: color.RGBA{R: 255, A: 255},
		FillColor:   color.RGBA{R: 255, A: 255},
		StrokeWidth: 5,
	})

	// Original should be unchanged
	after := pi.ToImage()
	if original.Pix[0] != after.Pix[0] {
		t.Error("original image was mutated by DrawRect")
	}
}
