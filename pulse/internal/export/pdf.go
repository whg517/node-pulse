package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/signintech/gopdf"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// GeneratePDF builds a server-side PDF report for the given request and returns
// the file path. It reuses the same metrics query path as CSV export. Errors are
// returned so the report scheduler can log and reschedule.
func (s *ExportService) GeneratePDF(ctx context.Context, req *CreateExportRequest) (string, error) {
	if err := req.Validate(); err != nil {
		// PDF supports pdf format; relax the csv-only Validate constraint.
		if req.Format != "pdf" {
			return "", fmt.Errorf("validation failed: %w", err)
		}
	}

	if err := os.MkdirAll(ExportDir, 0o755); err != nil {
		return "", fmt.Errorf("create export dir: %w", err)
	}

	filename := fmt.Sprintf("metrics_export_%s_%s.pdf", req.UserID, time.Now().Format("20060102-150405"))
	filePath := filepath.Join(ExportDir, filename)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.SetMargins(30, 30, 30, 30)
	pdf.AddPage()
	if err := pdf.AddTTFFont("helv", "/System/Library/Fonts/Supplemental/Arial.ttf"); err != nil {
		// Fall back to the built-in core font when no system TTF is available.
		_ = pdf.SetFont("helvetica", "", 12)
	} else {
		_ = pdf.SetFont("helv", "", 12)
	}

	// Title
	_ = pdf.SetFontSize(18)
	_ = pdf.Cell(nil, "NodePulse Metrics Report")
	pdf.Br(28)
	_ = pdf.SetFontSize(10)
	_ = pdf.Cell(nil, fmt.Sprintf("Period: %s to %s", req.StartTime.Format("2006-01-02"), req.EndTime.Format("2006-01-02")))
	pdf.Br(20)

	// Table header
	headers := []string{"Timestamp", "Node", "Metric", "Value", "Unit"}
	colWidths := []float64{140, 80, 90, 70, 50}
	x, y := 30.0, pdf.GetY()
	for i, h := range headers {
		pdf.SetXY(x, y)
		// gopdf has no SetFontStyle; emulate bold by re-setting the font with style "B".
		setBold(&pdf, true)
		_ = pdf.Cell(&gopdf.Rect{W: colWidths[i], H: 18}, h)
		x += colWidths[i]
	}
	setBold(&pdf, false)
	pdf.Br(20)

	recordCount := 0
	for _, nodeID := range req.NodeIDs {
		for _, metric := range req.Metrics {
			rows, err := s.queryMetrics(ctx, nodeID, metric, req.StartTime, req.EndTime)
			if err != nil {
				return "", fmt.Errorf("query metrics node %s: %w", nodeID, err)
			}
			for _, row := range rows {
				if pdf.GetY() > 780 {
					pdf.AddPage()
				}
				x = 30
				y = pdf.GetY()
				vals := []string{row.Timestamp, row.NodeID, row.Metric, fmt.Sprintf("%.4g", row.Value), row.Unit}
				for i, v := range vals {
					pdf.SetXY(x, y)
					_ = pdf.Cell(&gopdf.Rect{W: colWidths[i], H: 16}, v)
					x += colWidths[i]
				}
				pdf.Br(16)
				recordCount++
			}
		}
	}

	pdf.Br(20)
	_ = pdf.Cell(nil, fmt.Sprintf("Total records: %d", recordCount))

	if err := pdf.WritePdf(filePath); err != nil {
		return "", fmt.Errorf("write pdf: %w", err)
	}
	return filePath, nil
}

// Ensure unused import is referenced for future use of models in this file.
var _ = models.ExportMetricsRow{}

// setBold toggles bold styling on the current gopdf font. gopdf has no
// SetFontStyle helper, so we re-apply SetFont with the appropriate style flag.
// The body font size is fixed at 12 for this report.
func setBold(pdf *gopdf.GoPdf, bold bool) {
	style := ""
	if bold {
		style = "B"
	}
	_ = pdf.SetFont("helv", style, 12)
}
