package posts

import (
	"errors"

	"kbtuspace-backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(authorID int, input models.CreatePostInput) (*models.Post, error) {
	post := &models.Post{
		AuthorID:  authorID,
		FacultyID: input.FacultyID,
		Title:     input.Title,
		Content:   input.Content,
		ImageURL:  input.ImageURL,
		IsPinned:  input.IsPinned,
		Capacity:  0,
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) GetAll(facultyID *int) ([]models.Post, error) {
	return s.repo.GetAll(facultyID)
}

func (s *Service) GetByID(id int) (*models.Post, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Update(id int, currentUserID int, role string, input models.UpdatePostInput) error {
	authorID, err := s.repo.GetAuthorID(id)
	if err != nil {
		return err
	}

	if role != "admin" && authorID != currentUserID {
		return errors.New("forbidden")
	}

	post := &models.Post{
		ID:        id,
		FacultyID: input.FacultyID,
		Title:     input.Title,
		Content:   input.Content,
		ImageURL:  input.ImageURL,
		IsPinned:  input.IsPinned,
	}

	return s.repo.Update(post)
}

func (s *Service) Delete(id int, currentUserID int, role string) error {
	authorID, err := s.repo.GetAuthorID(id)
	if err != nil {
		return err
	}

	if role != "admin" && authorID != currentUserID {
		return errors.New("forbidden")
	}

	return s.repo.Delete(id)
}
