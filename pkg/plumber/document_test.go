package plumber

import (
	"testing"
)

const (
	testSimplePDF    = "../../testdata/simple.pdf"
	testTablePDF     = "../../testdata/simple_table.pdf"
	testMultipagePDF = "../../testdata/multipage.pdf"
)

func TestOpenAndClose(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	if doc.NumPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

func TestOpenNonExistent(t *testing.T) {
	_, err := Open("/nonexistent.pdf")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestDoubleClose(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := doc.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}

	if err := doc.Close(); err != ErrClosed {
		t.Errorf("second Close = %v, want ErrClosed", err)
	}
}

func TestPageOutOfRange(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	_, err = doc.Page(0)
	if err == nil {
		t.Error("expected error for page 0")
	}

	_, err = doc.Page(999)
	if err == nil {
		t.Error("expected error for page 999")
	}
}

func TestPageAccess(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	if page.Number() != 1 {
		t.Errorf("Number = %d, want 1", page.Number())
	}

	if page.Width() <= 0 {
		t.Errorf("Width = %f, expected > 0", page.Width())
	}
	if page.Height() <= 0 {
		t.Errorf("Height = %f, expected > 0", page.Height())
	}

	t.Logf("Page 1: %.0f x %.0f", page.Width(), page.Height())
}

func TestMultipage(t *testing.T) {
	doc, err := Open(testMultipagePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	if doc.NumPages() != 3 {
		t.Fatalf("NumPages = %d, want 3", doc.NumPages())
	}

	pages, err := doc.Pages()
	if err != nil {
		t.Fatalf("Pages: %v", err)
	}

	if len(pages) != 3 {
		t.Fatalf("len(Pages) = %d, want 3", len(pages))
	}

	for i, p := range pages {
		if p.Number() != i+1 {
			t.Errorf("page %d: Number = %d", i+1, p.Number())
		}
	}
}

func TestMetadata(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	meta := doc.Metadata()
	t.Logf("Metadata: %v", meta)
	// gopdf sets Creator and Producer
}

func TestPDFVersion(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	v := doc.PDFVersion()
	if v == "" {
		t.Error("expected non-empty PDF version")
	}
	t.Logf("PDF version: %s", v)
}

func TestClosedDocumentAccess(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	doc.Close()

	_, err = doc.Page(1)
	if err != ErrClosed {
		t.Errorf("Page on closed doc = %v, want ErrClosed", err)
	}

	_, err = doc.Pages()
	if err != ErrClosed {
		t.Errorf("Pages on closed doc = %v, want ErrClosed", err)
	}
}
