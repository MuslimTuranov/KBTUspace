package events

import (
	"strings"
	"time"

	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/cache"
)

func parseEventDate(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, ErrInvalidEventDate
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"02.01.2006",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, ErrInvalidEventDate
}

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

func eventsListKey(facultyID *int, role string, globalOnly bool) string {
	if globalOnly {
		return cache.EventsListPrefix() + "global_only"
	}
	return cache.EventsListKey(facultyID)
}

func canAccessEvent(role string, actorFacultyID *int, event *models.Post) bool {
	if role == "admin" {
		return true
	}
	return event.Status == models.ContentStatusApproved
}

func (s *Service) Create(authorID int, role string, actorFacultyID *int, input models.CreateEventInput) (*models.Post, error) {
	facultyID, scope, status, approvedBy, approvedAt, rejectionReason, err := resolveModeration(role, actorFacultyID, input.FacultyID, input.Scope)
	if err != nil {
		return nil, err
	}

	eventDate, err := parseEventDate(input.EventDate)
	if err != nil {
		return nil, err
	}
	if eventDate.Before(time.Now()) {
		return nil, ErrEventDateInPast
	}
	location := input.Location
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
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return post, nil
}

func (s *Service) GetAll(facultyID *int, role string, globalOnly bool) ([]models.Post, error) {
	if s.cache != nil {
		if cachedEvents, hit, err := s.cache.GetPosts(eventsListKey(facultyID, role, globalOnly)); err == nil && hit {
			return cachedEvents, nil
		}
	}

	events, err := s.repo.GetAll(facultyID, role, globalOnly)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPosts(eventsListKey(facultyID, role, globalOnly), events)
	}

	return events, nil
}

func (s *Service) GetByID(id int, userID int, role string, actorFacultyID *int) (*models.Post, error) {
	if s.cache != nil && role != "admin" {
		if cachedEvent, hit, err := s.cache.GetPost(eventKey(id)); err == nil && hit {
			if canAccessEvent(role, actorFacultyID, cachedEvent) {
				registered, err := s.repo.IsUserRegistered(userID, id)
				if err != nil {
					return nil, err
				}
				cachedEvent.IsRegistered = &registered
				return cachedEvent, nil
			}
		}
	}

	event, err := s.repo.GetByID(id, role == "admin", actorFacultyID)
	if err != nil {
		return nil, err
	}

	registered, err := s.repo.IsUserRegistered(userID, id)
	if err != nil {
		return nil, err
	}
	event.IsRegistered = &registered

	if s.cache != nil && event.Status == models.ContentStatusApproved {
		cacheEvent := *event
		cacheEvent.IsRegistered = nil
		_ = s.cache.SetPost(eventKey(id), &cacheEvent)
	}

	return event, nil
}

func (s *Service) Update(id int, actorID int, role string, actorFacultyID *int, input models.UpdateEventInput) error {
	facultyID, scope, status, approvedBy, approvedAt, rejectionReason, err := resolveModeration(role, actorFacultyID, input.FacultyID, input.Scope)
	if err != nil {
		return err
	}

	eventDate, err := parseEventDate(input.EventDate)
	if err != nil {
		return err
	}
	if eventDate.Before(time.Now()) {
		return ErrEventDateInPast
	}
	location := input.Location
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

	if role == "admin" && status == models.ContentStatusApproved {
		post.ApprovedBy = &actorID
	}

	if err := s.repo.Update(post, actorID, role == "admin", actorFacultyID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) Delete(id int, actorID int, role string, actorFacultyID *int) error {
	if err := s.repo.Delete(id, actorID, role == "admin", actorFacultyID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) Register(userID int, eventID int, role string, actorFacultyID *int) error {
	if err := s.repo.Register(userID, eventID, actorFacultyID, role == "admin"); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(eventID))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) CancelRegistration(userID, eventID int) error {
	if err := s.repo.CancelRegistration(userID, eventID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(eventID))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) MarkAttended(actorID int, actorRole string, actorFacultyID *int, userID, eventID int) error {
	meta, err := s.repo.GetEventAccessMeta(eventID)
	if err != nil {
		return err
	}

	if actorRole != "admin" {
		isOwner := meta.AuthorID == actorID
		sameFaculty := actorFacultyID != nil && meta.FacultyID != nil && *actorFacultyID == *meta.FacultyID
		if !isOwner && !sameFaculty {
			return ErrForbidden
		}
	}

	return s.repo.MarkAttended(userID, eventID)
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
