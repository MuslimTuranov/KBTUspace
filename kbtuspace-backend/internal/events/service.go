package events

import "kbtuspace-backend/internal/models"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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

	return event, nil
}

func (s *Service) GetAll(facultyID *int) ([]models.Post, error) {
	return s.repo.GetAll(facultyID)
}

func (s *Service) GetByID(id int) (*models.Post, error) {
	return s.repo.GetByID(id)
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

	return s.repo.Update(event)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}
