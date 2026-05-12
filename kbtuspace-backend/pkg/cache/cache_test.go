package cache

import (
	"testing"
	"time"

	"kbtuspace-backend/internal/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func intPtr(v int) *int {
	return &v
}

func newTestRedisCache(t *testing.T) (*redisCache, *miniredis.Miniredis) {
	t.Helper()

	s := miniredis.RunT(t)

	c := &redisCache{
		client: redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		}),
		expires: time.Minute,
	}

	return c, s
}

func TestCacheKeys(t *testing.T) {
	facultyID := 2

	if PostKey(1) != "posts:item:1" {
		t.Fatalf("wrong PostKey: %s", PostKey(1))
	}

	if PostsListKey(&facultyID) != "posts:list:faculty:2" {
		t.Fatalf("wrong PostsListKey: %s", PostsListKey(&facultyID))
	}

	if EventKey(1) != "events:item:1" {
		t.Fatalf("wrong EventKey: %s", EventKey(1))
	}

	if EventsListKey(&facultyID) != "events:list:faculty:2" {
		t.Fatalf("wrong EventsListKey: %s", EventsListKey(&facultyID))
	}
}

func TestNilFacultyIDGivesAll(t *testing.T) {
	if PostsListKey(nil) != "posts:list:all" {
		t.Fatalf("expected posts:list:all, got %s", PostsListKey(nil))
	}

	if EventsListKey(nil) != "events:list:all" {
		t.Fatalf("expected events:list:all, got %s", EventsListKey(nil))
	}
}

func TestSetPostGetPost(t *testing.T) {
	c, _ := newTestRedisCache(t)

	post := &models.Post{
		ID:       1,
		AuthorID: 10,
		Title:    "Post title",
		Content:  "Post content",
		Scope:    models.ContentScopeFaculty,
		Status:   models.ContentStatusApproved,
	}

	err := c.SetPost(PostKey(post.ID), post)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, ok, err := c.GetPost(PostKey(post.ID))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ok {
		t.Fatal("expected cache hit")
	}

	if got.ID != post.ID || got.Title != post.Title {
		t.Fatal("cached post mismatch")
	}
}

func TestSetPostsGetPosts(t *testing.T) {
	c, _ := newTestRedisCache(t)

	posts := []models.Post{
		{
			ID:       1,
			Title:    "Post 1",
			Content:  "Content 1",
			Scope:    models.ContentScopeFaculty,
			Status:   models.ContentStatusApproved,
			AuthorID: 1,
		},
		{
			ID:       2,
			Title:    "Post 2",
			Content:  "Content 2",
			Scope:    models.ContentScopeGlobal,
			Status:   models.ContentStatusApproved,
			AuthorID: 2,
		},
	}

	key := PostsListKey(intPtr(1))

	err := c.SetPosts(key, posts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, ok, err := c.GetPosts(key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ok {
		t.Fatal("expected cache hit")
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(got))
	}
}

func TestCacheMiss(t *testing.T) {
	c, _ := newTestRedisCache(t)

	post, ok, err := c.GetPost(PostKey(999))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ok {
		t.Fatal("expected cache miss")
	}

	if post != nil {
		t.Fatal("expected nil post")
	}

	posts, ok, err := c.GetPosts(PostsListKey(nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ok {
		t.Fatal("expected cache miss")
	}

	if posts != nil {
		t.Fatal("expected nil posts")
	}
}

func TestInvalidJSONInRedis(t *testing.T) {
	c, s := newTestRedisCache(t)

	s.Set(PostKey(1), "{invalid-json")

	_, ok, err := c.GetPost(PostKey(1))
	if err == nil {
		t.Fatal("expected invalid json error")
	}

	if ok {
		t.Fatal("expected ok false")
	}

	s.Set(PostsListKey(nil), "{invalid-json")

	_, ok, err = c.GetPosts(PostsListKey(nil))
	if err == nil {
		t.Fatal("expected invalid json error")
	}

	if ok {
		t.Fatal("expected ok false")
	}
}

func TestDelete(t *testing.T) {
	c, _ := newTestRedisCache(t)

	post := &models.Post{
		ID:       1,
		Title:    "Post",
		Content:  "Content",
		Scope:    models.ContentScopeFaculty,
		Status:   models.ContentStatusApproved,
		AuthorID: 1,
	}

	key := PostKey(1)

	err := c.SetPost(key, post)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = c.Delete(key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, ok, err := c.GetPost(key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ok {
		t.Fatal("expected deleted key")
	}
}

func TestDeletePrefixDeletesOnlyMatchingPrefix(t *testing.T) {
	c, _ := newTestRedisCache(t)

	post := &models.Post{
		ID:       1,
		Title:    "Post",
		Content:  "Content",
		Scope:    models.ContentScopeFaculty,
		Status:   models.ContentStatusApproved,
		AuthorID: 1,
	}

	err := c.SetPost(PostKey(1), post)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = c.SetPost(EventKey(1), post)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = c.DeletePrefix("posts:")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, postExists, err := c.GetPost(PostKey(1))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if postExists {
		t.Fatal("expected post key deleted")
	}

	_, eventExists, err := c.GetPost(EventKey(1))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !eventExists {
		t.Fatal("expected event key to remain")
	}
}
