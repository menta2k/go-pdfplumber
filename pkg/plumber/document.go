package plumber

import (
	"fmt"
	"io"
	"os"

	"github.com/digitorus/pdf"
)

// Document represents an opened PDF file.
type Document struct {
	reader   *pdf.Reader
	file     *os.File // non-nil only when opened via Open()
	closed   bool
	metadata map[string]string
}

// Open opens a PDF file at the given path and returns a Document.
func Open(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("plumber: open %s: %w", path, err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("plumber: stat %s: %w", path, err)
	}

	reader, err := pdf.NewReader(f, info.Size())
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("plumber: parse %s: %w", path, err)
	}

	return &Document{
		reader: reader,
		file:   f,
	}, nil
}

// OpenReader creates a Document from an io.ReaderAt and its size.
func OpenReader(r io.ReaderAt, size int64) (*Document, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("plumber: parse reader: %w", err)
	}

	return &Document{
		reader: reader,
	}, nil
}

// OpenEncrypted opens a password-protected PDF file.
func OpenEncrypted(path string, password string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("plumber: open %s: %w", path, err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("plumber: stat %s: %w", path, err)
	}

	called := false
	reader, err := pdf.NewReaderEncrypted(f, info.Size(), func() string {
		if called {
			return ""
		}
		called = true
		return password
	})
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("plumber: parse encrypted %s: %w", path, err)
	}

	return &Document{
		reader: reader,
		file:   f,
	}, nil
}

// Close releases all resources held by the document.
func (d *Document) Close() error {
	if d.closed {
		return ErrClosed
	}
	d.closed = true
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}

// NumPages returns the total number of pages.
func (d *Document) NumPages() int {
	return d.reader.NumPage()
}

// Page returns the page at the given 1-based index.
func (d *Document) Page(number int) (*Page, error) {
	if d.closed {
		return nil, ErrClosed
	}
	if number < 1 || number > d.reader.NumPage() {
		return nil, fmt.Errorf("%w: %d (document has %d pages)", ErrPageOutOfRange, number, d.reader.NumPage())
	}

	pdfPage := d.reader.Page(number)
	return newPageFromPDF(pdfPage, number), nil
}

// Pages returns all pages in the document.
func (d *Document) Pages() ([]*Page, error) {
	if d.closed {
		return nil, ErrClosed
	}

	n := d.reader.NumPage()
	pages := make([]*Page, 0, n)
	for i := 1; i <= n; i++ {
		p, err := d.Page(i)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, nil
}

// Metadata returns PDF document metadata (Title, Author, Subject, etc.).
func (d *Document) Metadata() map[string]string {
	if d.metadata != nil {
		return d.metadata
	}

	d.metadata = make(map[string]string)
	info := d.reader.Trailer().Key("Info")
	if info.IsNull() {
		return d.metadata
	}

	keys := []string{"Title", "Author", "Subject", "Keywords", "Creator", "Producer"}
	for _, k := range keys {
		v := info.Key(k)
		if !v.IsNull() {
			d.metadata[k] = v.Text()
		}
	}
	return d.metadata
}

// PDFVersion returns the PDF version string (e.g. "1.4", "2.0").
func (d *Document) PDFVersion() string {
	return d.reader.PDFVersion
}
