package cache

import (
	"fmt"

	"kbtuspace-backend/internal/models"
)

const (
	postKeyPrefix       = "posts:item:"
	postsListKeyPrefix  = "posts:list:"
	eventKeyPrefix      = "events:item:"
	eventsListKeyPrefix = "events:list:"
)

func PostKey(id int) string {
	return fmt.Sprintf("%s%d", postKeyPrefix, id)
}

func PostsListKey(facultyID *int) string {
	if facultyID == nil {
		return postsListKeyPrefix + "all"
	}

	return fmt.Sprintf("%sfaculty:%d", postsListKeyPrefix, *facultyID)
}

func PostsListPrefix() string {
	return postsListKeyPrefix
}

func EventKey(id int) string {
	return fmt.Sprintf("%s%d", eventKeyPrefix, id)
}

func EventsListKey(facultyID *int) string {
	if facultyID == nil {
		return eventsListKeyPrefix + "all"
	}

	return fmt.Sprintf("%sfaculty:%d", eventsListKeyPrefix, *facultyID)
}

func EventsListPrefix() string {
	return eventsListKeyPrefix
}

type PostsCache interface {
	SetPost(key string, value *models.Post) error
	GetPost(key string) (*models.Post, bool, error)
	SetPosts(key string, value []models.Post) error
	GetPosts(key string) ([]models.Post, bool, error)
	Delete(key string) error
	DeletePrefix(prefix string) error
}
