package events

import (
	"time"

	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/cache"
)

type Service struct {
	repo  *Repository
	cache cache.PostsCache
}

func NewService(repo *Repository, eventsCache cache.PostsCache) *Service {
	return &Service{repo: repo, cache: eventsCache}
}

func resolveModeration(role string, actorFacultyID *int, requestedFacultyID *int, requestedScope string) (*int, string, string, *int, *time.Time, *string, error) {
	scope := requestedScope
	if scope == "" {
		scope = models.ContentScopeFaculty
	}

	switch scope {
	case models.ContentScopeGlobal:
		if role == "admin" {
			now := time.Now()
			return nil, models.ContentScopeGlobal, models.ContentStatusApproved, nil, &now, nil, nil
		}
		if role == "organizer" {
			return nil, models.ContentScopeGlobal, models.ContentStatusPending, nil, nil, nil, nil
		}
		return nil, "", "", nil, nil, nil, ErrForbidden
	case models.ContentScopeFaculty:
		if role == "admin" && requestedFacultyID != nil {
			return requestedFacultyID, models.ContentScopeFaculty, models.ContentStatusApproved, nil, nil, nil, nil
		}
		if actorFacultyID == nil || *actorFacultyID <= 0 {
			return nil, "", "", nil, nil, nil, ErrFacultyRequired
		}
		return actorFacultyID, models.ContentScopeFaculty, models.ContentStatusApproved, nil, nil, nil, nil
	default:
		return nil, "", "", nil, nil, nil, ErrForbidden
	}
}

func eventKey(id int) string {
	return cache.EventKey(id)
}

func eventsListKey(facultyID *int) string {
	return cache.EventsListKey(facultyID)
}

func (s *Service) Create(authorID int, role string, actorFacultyID *int, input models.CreateEventInput) (*models.Post, error) {
	facultyID, scope, status, approvedBy, approvedAt, rejectionReason, err := resolveModeration(role, actorFacultyID, input.FacultyID, input.Scope)
	if err != nil {
		return nil, err
	}

	location := input.Location
	eventDate := input.EventDate
	post := &models.Post{
		AuthorID:        authorID,
		FacultyID:       facultyID,
		Title:           input.Title,
		Content:         input.Description,
		ImageURL:        input.ImageURL,
		Scope:           scope,
		Status:          status,
		ApprovedBy:      approvedBy,
		ApprovedAt:      approvedAt,
		RejectionReason: rejectionReason,
		Capacity:        input.Capacity,
		EventDate:       &eventDate,
		Location:        &location,
	}

	if role == "admin" && status == models.ContentStatusApproved {
		post.ApprovedBy = &authorID
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	if s.cache != nil && post.Status == models.ContentStatusApproved {
		_ = s.cache.SetPost(eventKey(post.ID), post)
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return post, nil
}

func (s *Service) GetAll(facultyID *int) ([]models.Post, error) {
	if s.cache != nil {
		if cachedEvents, hit, err := s.cache.GetPosts(eventsListKey(facultyID)); err == nil && hit {
			return cachedEvents, nil
		}
	}

	events, err := s.repo.GetAll(facultyID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPosts(eventsListKey(facultyID), events)
	}

	return events, nil
}

func (s *Service) GetByID(id int, role string) (*models.Post, error) {
	if s.cache != nil && role != "admin" {
		if cachedEvent, hit, err := s.cache.GetPost(eventKey(id)); err == nil && hit {
			return cachedEvent, nil
		}
	}

	event, err := s.repo.GetByID(id, role == "admin")
	if err != nil {
		return nil, err
	}

	if s.cache != nil && event.Status == models.ContentStatusApproved {
		_ = s.cache.SetPost(eventKey(id), event)
	}

	return event, nil
}

func (s *Service) Update(id int, role string, actorFacultyID *int, input models.UpdateEventInput) error {
	facultyID, scope, status, approvedBy, approvedAt, rejectionReason, err := resolveModeration(role, actorFacultyID, input.FacultyID, input.Scope)
	if err != nil {
		return err
	}

	location := input.Location
	eventDate := input.EventDate
	post := &models.Post{
		ID:              id,
		FacultyID:       facultyID,
		Title:           input.Title,
		Content:         input.Description,
		ImageURL:        input.ImageURL,
		Scope:           scope,
		Status:          status,
		ApprovedBy:      approvedBy,
		ApprovedAt:      approvedAt,
		RejectionReason: rejectionReason,
		IsPinned:        false,
		EventDate:       &eventDate,
		Location:        &location,
		Capacity:        input.Capacity,
	}

	if err := s.repo.Update(post, role == "admin", actorFacultyID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) Delete(id int, role string, actorFacultyID *int) error {
	if role != "admin" && (actorFacultyID == nil || *actorFacultyID <= 0) {
		return ErrFacultyRequired
	}

	if err := s.repo.Delete(id, role == "admin", actorFacultyID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) Register(userID int, eventID int) error {
	if err := s.repo.Register(userID, eventID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(eventID))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) ListPendingGlobal() ([]models.Post, error) {
	return s.repo.ListPendingGlobal()
}

func (s *Service) Approve(id int, adminID int) error {
	if err := s.repo.Approve(id, adminID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) Reject(id int, reason string) error {
	if err := s.repo.Reject(id, reason); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}
