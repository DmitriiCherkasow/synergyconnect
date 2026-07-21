package domain

import (
	"testing"
	"github.com/google/uuid"
)

func TestPost_IsAuthor(t *testing.T) {
	authorID := uuid.New()
	otherID := uuid.New()

	post := &Post{
		AuthorID: authorID,
	}

	if !post.IsAuthor(authorID) {
		t.Error("Expected true for author")
	}
	if post.IsAuthor(otherID) {
		t.Error("Expected false for non-author")
	}
}

func TestPost_CanView(t *testing.T) {
	authorID := uuid.New()
	otherID := uuid.New()
	groupID := uuid.New()

	// Публичный пост
	publicPost := &Post{
		AuthorID:   authorID,
		Visibility: VisibilityPublic,
	}
	if !publicPost.CanView(otherID, RoleStudent, []uuid.UUID{}) {
		t.Error("Public post should be visible to everyone")
	}

	// Приватный пост
	privatePost := &Post{
		AuthorID:   authorID,
		Visibility: VisibilityPrivate,
	}
	if !privatePost.CanView(authorID, RoleStudent, []uuid.UUID{}) {
		t.Error("Author should see their private post")
	}
	if privatePost.CanView(otherID, RoleStudent, []uuid.UUID{}) {
		t.Error("Others should not see private post")
	}

	// Групповой пост
	groupPost := &Post{
		AuthorID:   authorID,
		GroupID:    &groupID,
		Visibility: VisibilityGroup,
	}
	if !groupPost.CanView(authorID, RoleStudent, []uuid.UUID{groupID}) {
		t.Error("Member should see group post")
	}
	if groupPost.CanView(otherID, RoleStudent, []uuid.UUID{}) {
		t.Error("Non-member should not see group post")
	}
}