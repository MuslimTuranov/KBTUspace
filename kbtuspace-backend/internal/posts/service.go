package posts

import (
	"time"

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

func postKey(id int) string {
	return cache.PostKey(id)
}

func postsListKey(facultyID *int) string {
	return cache.PostsListKey(facultyID)
}

func (s *Service) Create(authorID int, role string, actorFacultyID *int, input models.CreatePostInput) (*models.Post, error) {
	facultyID, scope, status, approvedBy, approvedAt, rejectionReason, err := resolveModeration(role, actorFacultyID, input.FacultyID, input.Scope)
	if err != nil {
		return nil, err
	}

	post := &models.Post{
		AuthorID:        authorID,
		FacultyID:       facultyID,
		Title:           input.Title,
		Content:         input.Content,
		ImageURL:        input.ImageURL,
		Scope:           scope,
		Status:          status,
		ApprovedBy:      approvedBy,
		ApprovedAt:      approvedAt,
		RejectionReason: rejectionReason,
	}

	if role == "admin" && status == models.ContentStatusApproved {
		post.ApprovedBy = &authorID
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	if s.cache != nil && post.Status == models.ContentStatusApproved {
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

func (s *Service) GetByID(id int, role string) (*models.Post, error) {
	if s.cache != nil && role != "admin" {
		if cachedPost, hit, err := s.cache.GetPost(postKey(id)); err == nil && hit {
			return cachedPost, nil
		}
	}

	post, err := s.repo.GetByID(id, role == "admin")
	if err != nil {
		return nil, err
	}

	if s.cache != nil && post.Status == models.ContentStatusApproved {
		_ = s.cache.SetPost(postKey(id), post)
	}

	return post, nil
}

func (s *Service) Update(id int, currentUserID int, role string, actorFacultyID *int, input models.UpdatePostInput) error {
	if input.IsPinned && role == "student" {
		return ErrPinForbidden
	}

	facultyID, scope, status, approvedBy, approvedAt, rejectionReason, err := resolveModeration(role, actorFacultyID, input.FacultyID, input.Scope)
	if err != nil {
		return err
	}

	post := &models.Post{
		ID:              id,
		FacultyID:       facultyID,
		Title:           input.Title,
		Content:         input.Content,
		ImageURL:        input.ImageURL,
		IsPinned:        input.IsPinned,
		Scope:           scope,
		Status:          status,
		ApprovedBy:      approvedBy,
		ApprovedAt:      approvedAt,
		RejectionReason: rejectionReason,
	}

	if role == "admin" && status == models.ContentStatusApproved {
		post.ApprovedBy = &currentUserID
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

func (s *Service) ListPendingGlobal() ([]models.Post, error) {
	return s.repo.ListPendingGlobal()
}

func (s *Service) Approve(id int, adminID int) error {
	if err := s.repo.Approve(id, adminID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(postKey(id))
		_ = s.cache.DeletePrefix(cache.PostsListPrefix())
	}

	return nil
}

func (s *Service) Reject(id int, reason string) error {
	if err := s.repo.Reject(id, reason); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(postKey(id))
		_ = s.cache.DeletePrefix(cache.PostsListPrefix())
	}

	return nil
}

func (s *Service) Pin(id int, role string, actorFacultyID *int, isPinned bool) error {
	if role != "organizer" && role != "admin" {
		return ErrPinForbidden
	}
	if role != "admin" && (actorFacultyID == nil || *actorFacultyID <= 0) {
		return ErrFacultyRequired
	}

	if err := s.repo.Pin(id, isPinned, role == "admin", actorFacultyID); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.Delete(postKey(id))
		_ = s.cache.DeletePrefix(cache.PostsListPrefix())
	}

	return nil
}
