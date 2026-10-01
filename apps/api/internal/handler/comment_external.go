package handler

import (
	"net/http"
	"time"

	"auth/internal/domainfilter"
	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type externalCommentHandler struct {
	repo    repository.CommentRepository
	domains *domainfilter.Filter
}

func NewExternalCommentHandler(repo repository.CommentRepository, domains *domainfilter.Filter) *externalCommentHandler {
	return &externalCommentHandler{repo: repo, domains: domains}
}

func (h *externalCommentHandler) RegisterRoutes(router chi.Router) {
	router.Get("/{type}/{subjectKey}", httpx.EH(h.list))
	router.Get("/{type}/{subjectKey}/{rootId}/reply", httpx.EH(h.listReplies))
	router.With(httpx.RequireMember).Post("/{type}/{subjectKey}", httpx.EH(h.create))
	router.With(httpx.RequireMember).Patch("/{type}/{commentId}", httpx.EH(h.update))
	router.With(httpx.RequireAccessToken).Delete("/{type}/{commentId}", httpx.EH(h.delete))
	router.With(httpx.RequireAdmin).Put("/{type}/{commentId}/status", httpx.EH(h.setStatus))
}

type externalCommentResponse struct {
	ID             int64     `json:"id"`
	SubjectKey     string    `json:"subjectKey"`
	RootID         *int64    `json:"rootId"`
	Content        string    `json:"content"`
	AuthorID       int64     `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Status         int16     `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	ReplyCount     int64     `json:"replyCount"`
}

func newExternalCommentThreadResponse(r *http.Request, value repository.CommentThread) externalCommentResponse {
	response := newExternalCommentResponse(r, value.Comment)
	response.ReplyCount = value.ReplyCount
	return response
}

func newExternalCommentResponse(r *http.Request, value repository.Comment) externalCommentResponse {
	return externalCommentResponse{
		ID:             value.ID,
		SubjectKey:     value.SubjectKey,
		RootID:         value.RootID,
		Content:        publicCommentContent(r, value),
		AuthorID:       value.AuthorID,
		AuthorUsername: value.AuthorUsername,
		Status:         value.Status,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
	}
}

func externalCommentSubjectType(r *http.Request) (int16, error) {
	switch chi.URLParam(r, "type") {
	case "novel":
		return repository.CommentSubjectNovel, nil
	default:
		return 0, httpx.BadRequest("不支持的外部资源类型")
	}
}

func externalCommentID(r *http.Request) (int64, error) {
	return httpx.ParseParamPositiveInt(r, "commentId")
}

func externalCommentSubjectKey(r *http.Request) (string, error) {
	subjectKey := chi.URLParam(r, "subjectKey")
	if !validText(subjectKey, 1, 255) {
		return "", httpx.BadRequest("subjectKey 长度必须为 1 到 255")
	}
	return subjectKey, nil
}

func (h *externalCommentHandler) list(w http.ResponseWriter, r *http.Request) error {
	subjectType, err := externalCommentSubjectType(r)
	if err != nil {
		return err
	}
	subjectKey, err := externalCommentSubjectKey(r)
	if err != nil {
		return err
	}
	pagination, err := parsePagination(r.URL.Query(), 20, 100)
	if err != nil {
		return err
	}
	total, items, err := h.repo.ListRoots(subjectType, subjectKey, pagination.Limit, pagination.Offset)
	if err != nil {
		return httpx.InternalError(err, "查询附属资源评论失败")
	}
	response := make([]externalCommentResponse, len(items))
	for i, item := range items {
		response[i] = newExternalCommentThreadResponse(r, item)
	}
	render.JSON(w, r, page[externalCommentResponse]{Total: total, Items: response})
	return nil
}

func (h *externalCommentHandler) listReplies(w http.ResponseWriter, r *http.Request) error {
	subjectType, err := externalCommentSubjectType(r)
	if err != nil {
		return err
	}
	subjectKey, err := externalCommentSubjectKey(r)
	if err != nil {
		return err
	}
	rootID, err := httpx.ParseParamPositiveInt(r, "rootId")
	if err != nil {
		return err
	}
	pagination, err := parsePagination(r.URL.Query(), 20, 100)
	if err != nil {
		return err
	}
	total, items, err := h.repo.ListReplies(subjectType, subjectKey, rootID, pagination.Limit, pagination.Offset)
	if repository.IsNotFound(err) {
		return httpx.NotFound("一级评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "查询附属资源评论回复失败")
	}
	response := make([]externalCommentResponse, len(items))
	for i, item := range items {
		response[i] = newExternalCommentResponse(r, item)
	}
	render.JSON(w, r, page[externalCommentResponse]{Total: total, Items: response})
	return nil
}

func (h *externalCommentHandler) create(w http.ResponseWriter, r *http.Request) error {
	subjectType, err := externalCommentSubjectType(r)
	if err != nil {
		return err
	}
	subjectKey, err := externalCommentSubjectKey(r)
	if err != nil {
		return err
	}
	input, err := httpx.Body[commentInput](r)
	if err != nil {
		return err
	}
	if err := validateComment(input, h.domains); err != nil {
		return err
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	comment, err := h.repo.Create(repository.CreateCommentInput{
		SubjectType:    subjectType,
		SubjectKey:     subjectKey,
		RootID:         input.RootID,
		Content:        input.Content,
		AuthorID:       principal.UserID,
		AuthorUsername: principal.Username,
		Attr:           "{}",
	})
	if repository.IsNotFound(err) {
		return httpx.NotFound("根评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "创建附属资源评论失败")
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, newExternalCommentResponse(r, *comment))
	return nil
}

func (h *externalCommentHandler) modifiableID(r *http.Request) (int16, int64, error) {
	subjectType, err := externalCommentSubjectType(r)
	if err != nil {
		return 0, 0, err
	}
	id, err := externalCommentID(r)
	if err != nil {
		return 0, 0, err
	}
	comment, err := h.repo.Find(subjectType, id)
	if repository.IsNotFound(err) {
		return 0, 0, httpx.NotFound("评论不存在")
	}
	if err != nil {
		return 0, 0, httpx.InternalError(err, "查询评论失败")
	}
	principal, _ := httpx.AuthenticatedPrincipal(r)
	if comment.AuthorID != principal.UserID && !principal.IsAdmin() {
		return 0, 0, httpx.Forbidden("只能修改自己的评论")
	}

	const modificationWindow = 20 * time.Minute
	if !principal.IsAdmin() && time.Now().After(comment.CreatedAt.Add(modificationWindow)) {
		return 0, 0, httpx.Forbidden("评论只能在发布后 20 分钟内编辑或删除")
	}
	return subjectType, comment.ID, nil
}

func (h *externalCommentHandler) update(w http.ResponseWriter, r *http.Request) error {
	subjectType, id, err := h.modifiableID(r)
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
	comment, err := h.repo.Update(subjectType, id, input.Content)
	if repository.IsNotFound(err) {
		return httpx.NotFound("评论不存在")
	}
	if err != nil {
		return httpx.InternalError(err, "更新评论失败")
	}
	render.JSON(w, r, newExternalCommentResponse(r, *comment))
	return nil
}

func (h *externalCommentHandler) delete(w http.ResponseWriter, r *http.Request) error {
	subjectType, id, err := h.modifiableID(r)
	if err != nil {
		return err
	}
	if err := h.repo.SetStatus(subjectType, id, repository.StatusDeleted); err != nil {
		if repository.IsNotFound(err) {
			return httpx.NotFound("评论不存在")
		}
		return httpx.InternalError(err, "删除评论失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *externalCommentHandler) setStatus(w http.ResponseWriter, r *http.Request) error {
	subjectType, err := externalCommentSubjectType(r)
	if err != nil {
		return err
	}
	id, err := externalCommentID(r)
	if err != nil {
		return err
	}
	input, err := httpx.Body[commentStatusInput](r)
	if err != nil {
		return err
	}
	var status int16
	switch input.Status {
	case "published":
		status = repository.StatusPublished
	case "hidden":
		status = repository.StatusHidden
	case "deleted":
		status = repository.StatusDeleted
	default:
		return httpx.BadRequest("status 必须为 published、hidden 或 deleted")
	}
	if err := h.repo.SetStatus(subjectType, id, status); err != nil {
		if repository.IsNotFound(err) {
			return httpx.NotFound("评论不存在")
		}
		return httpx.InternalError(err, "设置评论状态失败")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
