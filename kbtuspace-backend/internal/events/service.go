package events

import (
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

func eventKey(id int) string {
	return cache.EventKey(id)
}

func eventsListKey(facultyID *int) string {
	return cache.EventsListKey(facultyID)
}

func (s *Service) Create(authorID int, input models.CreateEventInput) (*models.Post, error) {
	location := input.Location
	eventDate := input.EventDate

	event := &models.Post{
		AuthorID:  authorID,
		FacultyID: input.FacultyID,
		Title:     input.Title,
		Content:   input.Content,
		ImageURL:  input.ImageURL,
		IsPinned:  input.IsPinned,
		EventDate: &eventDate,
		Location:  &location,
		Capacity:  input.Capacity,
	}

	if err := s.repo.Create(event); err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPost(eventKey(event.ID), event)
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return event, nil
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

func (s *Service) GetByID(id int) (*models.Post, error) {
	if s.cache != nil {
		if cachedEvent, hit, err := s.cache.GetPost(eventKey(id)); err == nil && hit {
			return cachedEvent, nil
		}
	}

	event, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPost(eventKey(id), event)
	}

	return event, nil
}

func (s *Service) Update(id int, input models.UpdateEventInput) error {
	location := input.Location
	eventDate := input.EventDate

	event := &models.Post{
		ID:        id,
		FacultyID: input.FacultyID,
		Title:     input.Title,
		Content:   input.Content,
		ImageURL:  input.ImageURL,
		IsPinned:  input.IsPinned,
		EventDate: &eventDate,
		Location:  &location,
		Capacity:  input.Capacity,
	}

	if err := s.repo.Update(event); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(eventKey(id))
		_ = s.cache.DeletePrefix(cache.EventsListPrefix())
	}

	return nil
}

func (s *Service) Delete(id int) error {
	if err := s.repo.Delete(id); err != nil {
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

