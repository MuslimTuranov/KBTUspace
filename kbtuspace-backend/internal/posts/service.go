package posts

import (
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

func resolveFacultyID(role string, actorFacultyID *int, requestedFacultyID *int) (*int, error) {
	if role == "admin" && requestedFacultyID != nil {
		return requestedFacultyID, nil
	}
	if actorFacultyID == nil || *actorFacultyID <= 0 {
		return nil, ErrFacultyRequired
	}
	return actorFacultyID, nil
}

func postKey(id int) string {
	return cache.PostKey(id)
}

func postsListKey(facultyID *int) string {
	return cache.PostsListKey(facultyID)
}

func (s *Service) Create(authorID int, role string, actorFacultyID *int, input models.CreatePostInput) (*models.Post, error) {
	facultyID, err := resolveFacultyID(role, actorFacultyID, input.FacultyID)
	if err != nil {
		return nil, err
	}

	post := &models.Post{
		AuthorID:  authorID,
		FacultyID: facultyID,
		Title:     input.Title,
		Content:   input.Content,
		ImageURL:  input.ImageURL,
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPost(postKey(post.ID), post)
		_ = s.cache.DeletePrefix(cache.PostsListPrefix())
	}

	return post, nil
}

func (s *Service) GetAll(facultyID *int) ([]models.Post, error) {
	if s.cache != nil {
		if cachedPosts, hit, err := s.cache.GetPosts(postsListKey(facultyID)); err == nil && hit {
			return cachedPosts, nil
		}
	}

	posts, err := s.repo.GetAll(facultyID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPosts(postsListKey(facultyID), posts)
	}

	return posts, nil
}

func (s *Service) GetByID(id int) (*models.Post, error) {
	if s.cache != nil {
		if cachedPost, hit, err := s.cache.GetPost(postKey(id)); err == nil && hit {
			return cachedPost, nil
		}
	}

	post, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetPost(postKey(id), post)
	}

	return post, nil
}

func (s *Service) Update(id int, currentUserID int, role string, actorFacultyID *int, input models.UpdatePostInput) error {
	if input.IsPinned && role == "student" {
		return ErrPinForbidden
	}

	facultyID, err := resolveFacultyID(role, actorFacultyID, input.FacultyID)
	if err != nil {
		return err
	}

	post := &models.Post{
		ID:        id,
		FacultyID: facultyID,
		Title:     input.Title,
		Content:   input.Content,
		ImageURL:  input.ImageURL,
		IsPinned:  input.IsPinned,
	}

	if err := s.repo.Update(post, currentUserID, role == "admin"); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(postKey(id))
		_ = s.cache.DeletePrefix(cache.PostsListPrefix())
	}

	return nil
}

func (s *Service) Delete(id int, currentUserID int, role string) error {
	if err := s.repo.Delete(id, currentUserID, role == "admin"); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(postKey(id))
		_ = s.cache.DeletePrefix(cache.PostsListPrefix())
	}

	return nil
}
