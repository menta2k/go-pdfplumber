package plumber

import (
	"fmt"
	imgcolor "image/color"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcolor "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/color"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// DebugOverlay describes a rectangle to draw on a debug PDF.
type DebugOverlay struct {
	BBox        BBox
	Label       string
	StrokeColor imgcolor.RGBA
	FillColor   imgcolor.RGBA
}

// SaveDebugPDF creates a copy of the source PDF with colored square annotations
// drawn on top of the specified page. Opens in any PDF viewer for debugging.
func SaveDebugPDF(sourcePDF string, pageNum int, overlays []DebugOverlay, outputPath string) error {
	conf := model.NewDefaultConfiguration()

	annots := make([]model.AnnotationRenderer, 0, len(overlays))
	for _, ov := range overlays {
		sc := ov.StrokeColor
		if sc.A == 0 {
			sc = imgcolor.RGBA{R: 255, A: 255}
		}

		strokeCol := &pdfcolor.SimpleColor{
			R: float32(sc.R) / 255,
			G: float32(sc.G) / 255,
			B: float32(sc.B) / 255,
		}

		rect := types.Rectangle{
			LL: types.Point{X: ov.BBox.X0, Y: ov.BBox.Y0},
			UR: types.Point{X: ov.BBox.X1, Y: ov.BBox.Y1},
		}

		ann := model.NewSquareAnnotation(
			rect,
			0,              // apObjNr
			ov.Label,       // contents
			"",             // id
			"",             // modDate
			model.AnnPrint, // flags: print
			strokeCol,      // color
			"",             // title
			nil,            // popupIndRef
			nil,            // ca (opacity)
			"",             // rc
			"",             // subject
			nil,            // fillCol
			0, 0, 0, 0,    // margins
			1.5,            // borderWidth
			model.BSSolid,  // borderStyle
			false,          // cloudyBorder
			0,              // cloudyBorderIntensity
		)

		annots = append(annots, ann)
	}

	m := map[int][]model.AnnotationRenderer{
		pageNum: annots,
	}

	if err := api.AddAnnotationsMapFile(sourcePDF, outputPath, m, conf, false); err != nil {
		return fmt.Errorf("add annotations: %w", err)
	}

	return nil
}
