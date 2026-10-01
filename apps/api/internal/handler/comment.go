package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	"auth/internal/domainfilter"
	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type commentHandler struct {
	repo    repository.CommentRepository
	domains *domainfilter.Filter
}

func NewCommentHandler(repo repository.CommentRepository, domains *domainfilter.Filter) *commentHandler {
	return &commentHandler{repo: repo, domains: domains}
}

func (h *commentHandler) RegisterRoutes(router chi.Router) {
	router.With(httpx.RequireMember).Patch("/{id}", httpx.EH(h.update))
	router.With(httpx.RequireAccessToken).Delete("/{id}", httpx.EH(h.delete))
}

type commentInput struct {
	Content string `json:"content"`
	RootID  *int64 `json:"rootId"`
}

type commentUpdateInput struct {
	Content string          `json:"content"`
	RootID  json.RawMessage `json:"rootId"`
}

func validateCommentUpdate(input commentUpdateInput, domains *domainfilter.Filter) error {
	if len(input.RootID) > 0 {
		return httpx.BadRequest("rootId 不能修改")
	}
	return validateComment(commentInput{Content: input.Content}, domains)
}

func validateComment(input commentInput, domains *domainfilter.Filter) error {
	if !validText(input.Content, 1, 1000) || utf8.RuneCountInString(input.Content) > 1000 {
		return httpx.BadRequest("content 不能为空且不能超过 1000 字")
	}
	if input.RootID != nil && *input.RootID <= 0 {
		return httpx.BadRequest("rootId 必须为正整数")
	}
	return checkDomainText(domains, "content", input.Content)
}

type commentResponse struct {
	ID             int64                  `json:"id"`
	PostID         int64                  `json:"postId"`
	RootID         *int64                 `json:"rootId"`
	Content        string                 `json:"content"`
	AuthorID       int64                  `json:"authorId"`
	AuthorUsername string                 `json:"authorUsername"`
	Status         int16                  `json:"status"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	ReplyCount     int64                  `json:"replyCount"`
	Replies        *page[commentResponse] `json:"replies,omitempty"`
}

func newCommentResponse(r *http.Request, value repository.Comment) (commentResponse, error) {
	postID, err := repository.PostIDFromSubjectKey(value.SubjectKey)
	if err != nil {
		return commentResponse{}, fmt.Errorf("comment %d: %w", value.ID, err)
	}
	return commentResponse{
		ID:             value.ID,
		PostID:         postID,
		RootID:         value.RootID,
		Content:        publicCommentContent(r, value),
		AuthorID:       value.AuthorID,
		AuthorUsername: value.AuthorUsername,
		Status:         value.Status,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
	}, nil
}

func newCommentThreadResponse(r *http.Request, value repository.CommentThread) (commentResponse, error) {
	response, err := newCommentResponse(r, value.Comment)
	if err != nil {
		return response, err
	}
	response.ReplyCount = value.ReplyCount
	items := make([]commentResponse, len(value.Replies))
	for i, reply := range value.Replies {
		items[i], err = newCommentResponse(r, reply)
		if err != nil {
			return response, err
		}
	}
	response.Replies = &page[commentResponse]{Total: value.ReplyCount, Items: items}
	return response, nil
}

func publicCommentContent(r *http.Request, value repository.Comment) string {
	principal, err := httpx.AuthenticatedPrincipal(r)
	if value.Status == repository.StatusPublished || (err == nil && principal.IsAdmin()) {
		return value.Content
	}
	return ""
}

func (h *commentHandler) modifiableID(r *http.Request) (int64, error) {
	id, err := httpx.ParseParamPositiveInt(r, "id")
	if err != nil {
		return 0, err
	}
	comment, err := h.repo.Find(repository.CommentSubjectPost, id)
	if repository.IsNotFound(err) {
		return 0, httpx.NotFound("评论不存在")
	}
	if err != nil {
		return 0, httpx.InternalError(err, "查询评论失败")
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if comment.AuthorID != principal.UserID && !principal.IsAdmin() {
		return 0, httpx.Forbidden("只能修改自己的评论")
	}

	const modificationWindow = 20 * time.Minute
	if !principal.IsAdmin() && time.Now().After(comment.CreatedAt.Add(modificationWindow)) {
		return 0, httpx.Forbidden("评论只能在发布后 20 分钟内编辑或删除")
	}
	return comment.ID, nil
}

func (h *commentHandler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := h.modifiableID(r)
	if err != nil {
		return err
	}
	input, err := httpx.Body[commentUpdateInput](r)
	if err != nil {
		return err
	}
	if err := validateCommentUpdate(input, h.domains); err != nil {
		return err
	}
	comment, err := h.repo.Update(repository.CommentSubjectPost, id, input.Content)
	if repository.IsNotFound(err) {
		return httpx.NotFound("评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "更新评论失败")
	}
	response, err := newCommentResponse(r, *comment)
	if err != nil {
		return httpx.InternalError(err, "转换评论数据失败")
	}
	render.JSON(w, r, response)
	return nil
}

func (h *commentHandler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := h.modifiableID(r)
	if err != nil {
		return err
	}
	err = h.repo.SetStatus(repository.CommentSubjectPost, id, repository.StatusDeleted)
	if repository.IsNotFound(err) {
		return httpx.NotFound("评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "删除评论失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
