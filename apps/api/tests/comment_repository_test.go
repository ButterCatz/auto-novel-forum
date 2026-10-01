//go:build integration

package tests

import (
	forumcategory "auth/internal/category"
	"auth/internal/repository"
	"testing"
)

func TestCommentRepositoryDeleteAllByAuthor(t *testing.T) {
	resetDatabase()
	category, _ := forumcategory.FindByID(forumcategory.NovelID)
	firstPost, err := postRepo.Create(repository.CreatePostInput{
		CategoryID: category.ID, Title: "第一篇帖子", Content: "正文",
		AuthorID: 1, AuthorUsername: "author", Attr: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	secondPost, err := postRepo.Create(repository.CreatePostInput{
		CategoryID: category.ID, Title: "第二篇帖子", Content: "正文",
		AuthorID: 1, AuthorUsername: "author", Attr: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	createComment := func(subjectType int16, subjectKey string, authorID int64) *repository.Comment {
		t.Helper()
		comment, err := commentRepo.Create(repository.CreateCommentInput{
			SubjectType: subjectType, SubjectKey: subjectKey, Content: "评论",
			AuthorID: authorID, AuthorUsername: "commenter", Attr: `{}`,
		})
		if err != nil {
			t.Fatal(err)
		}
		return comment
	}

	firstPublished := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(firstPost.ID), 7)
	firstHidden := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(firstPost.ID), 7)
	if err := commentRepo.SetStatus(repository.CommentSubjectPost, firstHidden.ID, repository.StatusHidden); err != nil {
		t.Fatal(err)
	}
	secondPublished := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(secondPost.ID), 7)
	external := createComment(repository.CommentSubjectNovel, "novel:chapter-1", 7)
	otherAuthor := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(firstPost.ID), 8)

	if err := commentRepo.DeleteAllByAuthor(7); err != nil {
		t.Fatal(err)
	}
	// 重复删除必须保持幂等，不能再次扣减帖子评论数。
	if err := commentRepo.DeleteAllByAuthor(7); err != nil {
		t.Fatal(err)
	}

	for subjectType, commentIDs := range map[int16][]int64{
		repository.CommentSubjectPost:  {firstPublished.ID, firstHidden.ID, secondPublished.ID},
		repository.CommentSubjectNovel: {external.ID},
	} {
		for _, commentID := range commentIDs {
			comment, err := commentRepo.Find(subjectType, commentID)
			if err != nil {
				t.Fatal(err)
			}
			if comment.Status != repository.StatusDeleted {
				t.Fatalf("comment %d status = %d", comment.ID, comment.Status)
			}
		}
	}

	remaining, err := commentRepo.Find(repository.CommentSubjectPost, otherAuthor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if remaining.Status != repository.StatusPublished {
		t.Fatalf("other author's comment status = %d", remaining.Status)
	}

	first, err := postRepo.Find(firstPost.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := postRepo.Find(secondPost.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if first.CommentsCount != 1 || second.CommentsCount != 0 {
		t.Fatalf("unexpected comment counts: first=%d second=%d", first.CommentsCount, second.CommentsCount)
	}
}

func TestCommentRootReplyPreviews(t *testing.T) {
	resetDatabase()
	create := func(key string, rootID *int64) *repository.Comment {
		t.Helper()
		c, err := commentRepo.Create(repository.CreateCommentInput{
			SubjectType: repository.CommentSubjectNovel, SubjectKey: key, RootID: rootID,
			Content: "reply", AuthorID: 1, AuthorUsername: "reader", Attr: "{}",
		})
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	roots := []*repository.Comment{create("preview", nil), create("preview", nil), create("preview", nil)}
	for _, root := range roots[:2] {
		for i := 0; i < 23; i++ {
			create("preview", &root.ID)
		}
	}
	other := create("other", nil)
	create("other", &other.ID)
	total, threads, err := commentRepo.ListRoots(repository.CommentSubjectNovel, "preview", 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(threads) != 3 {
		t.Fatalf("unexpected roots: %d %d", total, len(threads))
	}
	for i, thread := range threads {
		if i == 2 {
			if thread.ReplyCount != 0 || len(thread.Replies) != 0 {
				t.Fatal("empty root has replies")
			}
			continue
		}
		count, replies, err := commentRepo.ListReplies(repository.CommentSubjectNovel, "preview", thread.ID, 20, 0)
		if err != nil {
			t.Fatal(err)
		}
		if thread.ReplyCount != count || count != 23 || len(thread.Replies) != 20 {
			t.Fatalf("bad preview size/count: %#v", thread)
		}
		for j, reply := range thread.Replies {
			if reply.ID != replies[j].ID || reply.RootID == nil || *reply.RootID != thread.ID {
				t.Fatalf("preview order/group mismatch: %#v", reply)
			}
		}
	}
	_, page, err := commentRepo.ListRoots(repository.CommentSubjectNovel, "preview", 1, 1)
	if err != nil || len(page) != 1 || page[0].ID != roots[1].ID || len(page[0].Replies) != 20 {
		t.Fatalf("root pagination: %#v %v", page, err)
	}
	_, empty, err := commentRepo.ListRoots(repository.CommentSubjectNovel, "preview", 3, 3)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty page: %#v %v", empty, err)
	}
}
