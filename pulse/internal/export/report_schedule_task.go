package export

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/notify"
)

// ReportScheduleRunner is a scheduler.Task that polls due report schedules,
// generates the export artifact (CSV via ExportService; PDF via GeneratePDF),
// emails it to the owner, and updates last/next run timestamps.
type ReportScheduleRunner struct {
	repo       reportScheduleStore
	exports    exportRunner
	mailer     notify.Sender
	userLookup userEmailer
	getExport  func(exportID string) (*models.ExportTask, error)
	interval   time.Duration
}

// reportScheduleStore is the subset of db.ReportScheduleRepository used here.
type reportScheduleStore interface {
	ListDue(ctx context.Context, now time.Time) ([]*models.ReportSchedule, error)
	MarkRun(ctx context.Context, id string, lastRun, nextRun time.Time) error
}

// exportRunner is the subset of ExportService used here.
type exportRunner interface {
	CreateExport(ctx context.Context, req *CreateExportRequest) (*models.ExportTask, error)
	GeneratePDF(ctx context.Context, req *CreateExportRequest) (string, error)
}

// userEmailer resolves a recipient email for a schedule (recipient_email or owner's email).
type userEmailer func(ctx context.Context, userID string) (string, error)

// NewReportScheduleRunner builds the task. interval defaults to 1 minute.
// getExport wires ExportService.GetExport for polling the async CSV path.
func NewReportScheduleRunner(
	repo reportScheduleStore,
	exports exportRunner,
	mailer notify.Sender,
	userLookup userEmailer,
	getExport func(exportID string) (*models.ExportTask, error),
) *ReportScheduleRunner {
	return &ReportScheduleRunner{repo: repo, exports: exports, mailer: mailer, userLookup: userLookup, getExport: getExport, interval: time.Minute}
}

// Name implements scheduler.Task.
func (r *ReportScheduleRunner) Name() string { return "report-schedule" }

// Interval implements scheduler.Task.
func (r *ReportScheduleRunner) Interval() time.Duration { return r.interval }

// Execute polls due schedules and processes each. Errors are logged, not fatal.
func (r *ReportScheduleRunner) Execute(ctx context.Context) error {
	if r.repo == nil || r.exports == nil {
		return nil
	}
	due, err := r.repo.ListDue(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("list due schedules: %w", err)
	}
	for _, sch := range due {
		r.processOne(ctx, sch)
	}
	return nil
}

func (r *ReportScheduleRunner) processOne(ctx context.Context, sch *models.ReportSchedule) {
	logger := slog.With("component", "report-schedule", "schedule_id", sch.ID, "name", sch.Name)

	end := time.Now()
	var start time.Time
	switch sch.Frequency {
	case "weekly":
		start = end.AddDate(0, 0, -7)
	case "monthly":
		start = end.AddDate(0, -1, 0)
	default:
		start = end.AddDate(0, 0, -1)
	}

	req := &CreateExportRequest{
		UserID:    sch.OwnerUserID,
		NodeIDs:   sch.NodeIDs,
		StartTime: start,
		EndTime:   end,
		Metrics:   sch.Metrics,
		Format:    sch.Format,
	}

	var (
		path string
		err  error
	)
	if sch.Format == "pdf" {
		path, err = r.exports.GeneratePDF(ctx, req)
	} else {
		// Default to CSV (the only plain export backend today).
		req.Format = "csv"
		var task *models.ExportTask
		// CreateExport runs async; poll briefly for the file path.
		task, err = r.exports.CreateExport(ctx, req)
		if err == nil {
			path = r.waitForPath(task.ID)
		}
	}
	if err != nil {
		logger.Error("Failed to generate scheduled report", "error", err)
		r.markNext(ctx, sch, time.Now())
		return
	}

	// Resolve recipient.
	to := sch.RecipientEmail
	if to == "" && r.userLookup != nil {
		if email, lookupErr := r.userLookup(ctx, sch.OwnerUserID); lookupErr == nil {
			to = email
		}
	}
	if to == "" {
		logger.Warn("No recipient for scheduled report; artifact kept on disk", "path", path)
	} else {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			logger.Error("Failed to read report artifact", "path", path, "error", readErr)
		} else {
			subject := fmt.Sprintf("NodePulse report — %s", sch.Name)
			body := fmt.Sprintf("Your scheduled report %q is attached.\n", sch.Name)
			ct := "text/csv"
			if sch.Format == "pdf" {
				ct = "application/pdf"
			}
			if sendErr := r.mailer.Send(ctx, to, subject, body, notify.Attachment{
				Filename: fmt.Sprintf("%s.%s", sch.Name, sch.Format), Content: content, ContentType: ct,
			}); sendErr != nil {
				logger.Error("Failed to email scheduled report", "to", to, "error", sendErr)
			}
		}
	}

	r.markNext(ctx, sch, time.Now())
}

// markNext records the run and schedules the next one based on frequency.
func (r *ReportScheduleRunner) markNext(ctx context.Context, sch *models.ReportSchedule, lastRun time.Time) {
	var next time.Time
	switch sch.Frequency {
	case "weekly":
		next = lastRun.AddDate(0, 0, 7)
	case "monthly":
		next = lastRun.AddDate(0, 1, 0)
	default:
		next = lastRun.AddDate(0, 0, 1)
	}
	if err := r.repo.MarkRun(ctx, sch.ID, lastRun, next); err != nil {
		slog.Error("Failed to mark schedule run", "component", "report-schedule", "schedule_id", sch.ID, "error", err)
	}
}

// waitForPath polls ExportService.GetExport until the async CSV job finishes
// (or a short timeout). Returns the file path or "" if unavailable.
func (r *ReportScheduleRunner) waitForPath(exportID string) string {
	if r.getExport == nil {
		return ""
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		task, err := r.getExport(exportID)
		if err == nil && task != nil && task.IsCompleted() && task.FilePath != "" {
			return task.FilePath
		}
		time.Sleep(500 * time.Millisecond)
	}
	slog.Warn("Timed out waiting for scheduled export to complete", "component", "report-schedule", "export_id", exportID)
	return ""
}
