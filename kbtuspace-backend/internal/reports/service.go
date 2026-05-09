package reports

import (
	"database/sql"

	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/cache"
)

type Service struct {
	repo  *Repository
	cache cache.PostsCache
}

func NewService(repo *Repository, postsCache cache.PostsCache) *Service {
	return &Service{repo: repo, cache: postsCache}
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
	target, err := s.repo.CloseAndDeleteTarget(id, input.Status, input.ReviewNote, adminID)
	if err != nil {
		return err
	}

	if s.cache != nil && input.Status == models.ReportStatusClosed {
		switch target.TargetType {
		case models.ReportTargetEvent:
			_ = s.cache.Delete(cache.EventKey(target.TargetPostID))
			_ = s.cache.DeletePrefix(cache.EventsListPrefix())
		default:
			_ = s.cache.Delete(cache.PostKey(target.TargetPostID))
			_ = s.cache.DeletePrefix(cache.PostsListPrefix())
		}
	}

	return nil
}
