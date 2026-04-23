package reports

import (
	"database/sql"

	"kbtuspace-backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(reporterID int, input models.CreateReportInput) (*models.Report, error) {
	target, err := s.repo.GetTarget(input.TargetID, input.TargetType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	if target.AuthorID == reporterID {
		return nil, ErrSelfReport
	}

	exists, err := s.repo.HasPendingDuplicate(reporterID, target.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicatePending
	}

	report := &models.Report{
		ReporterID:     reporterID,
		TargetPostID:   target.ID,
		TargetType:     input.TargetType,
		Reason:         input.Reason,
		Status:         models.ReportStatusPending,
		TargetTitle:    target.Title,
		TargetAuthorID: target.AuthorID,
	}

	if err := s.repo.Create(report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *Service) List(status string) ([]models.Report, error) {
	if status == "" {
		status = models.ReportStatusPending
	}

	switch status {
	case models.ReportStatusPending, models.ReportStatusClosed, models.ReportStatusRejected:
	default:
		status = models.ReportStatusPending
	}

	return s.repo.List(status)
}

func (s *Service) Close(id, adminID int, input models.CloseReportInput) error {
	return s.repo.Close(id, input.Status, input.ReviewNote, adminID)
}
