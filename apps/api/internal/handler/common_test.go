package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth/internal/domainfilter"
	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type domainCommentRepository struct {
	repository.CommentRepository
	written bool
}

func (r *domainCommentRepository) Find(subjectType int16, id int64) (*repository.Comment, error) {
	return &repository.Comment{ID: id, AuthorID: 1, SubjectType: subjectType, CreatedAt: time.Now()}, nil
}

func (r *domainCommentRepository) Create(repository.CreateCommentInput) (*repository.Comment, error) {
	r.written = true
	return &repository.Comment{}, nil
}

func (r *domainCommentRepository) Update(int16, int64, string) (*repository.Comment, error) {
	r.written = true
	return &repository.Comment{}, nil
}

func TestDomainFilterRejectsWrites(t *testing.T) {
	domains, err := domainfilter.New([]domainfilter.Rule{{Domain: "evil.example", IncludeSubdomains: true}})
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "tester", "uid": 1, "role": "member",
	}).SignedString([]byte(httpx.AccessTokenSecret))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, method, path, body string
	}{
		{"post title create", http.MethodPost, "/post/", `{"categoryId":1,"title":"evil.example","content":"正文"}`},
		{"post content create", http.MethodPost, "/post/", `{"categoryId":1,"title":"标题","content":"evil.example"}`},
		{"post title update", http.MethodPatch, "/post/42/", `{"categoryId":1,"title":"evil.example","content":"正文"}`},
		{"post content update", http.MethodPatch, "/post/42/", `{"categoryId":1,"title":"标题","content":"evil.example"}`},
		{"post comment create", http.MethodPost, "/post/42/comment", `{"content":"evil.example"}`},
		{"post comment update", http.MethodPatch, "/comment/7", `{"content":"evil.example"}`},
		{"external comment create", http.MethodPost, "/external/comment/novel/book", `{"content":"evil.example"}`},
		{"external comment update", http.MethodPatch, "/external/comment/novel/7", `{"content":"evil.example"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			posts := &writePostRepository{}
			comments := &domainCommentRepository{}
			router := chi.NewRouter()
			router.Route("/post", NewPostHandler(posts, noFavoriteRepository{}, comments, domains).RegisterRoutes)
			router.Route("/comment", NewCommentHandler(comments, domains).RegisterRoutes)
			router.Route("/external/comment", NewExternalCommentHandler(comments, domains, nil).RegisterRoutes)
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || posts.written || comments.written {
				t.Fatalf("status=%d postWritten=%v commentWritten=%v body=%s", response.Code, posts.written, comments.written, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), "禁止使用的域名") {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
		})
	}
}

func TestDomainFilterErrorMapping(t *testing.T) {
	if err := checkDomainText(nil, "content", strings.Repeat("a", domainfilter.MaxTextBytes+1)); err != nil {
		t.Fatalf("disabled filter rejected content: %v", err)
	}
	domains, err := domainfilter.New([]domainfilter.Rule{{Domain: "evil.example", IncludeSubdomains: true}})
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{strings.Repeat("a", domainfilter.MaxTextBytes+1), strings.Repeat("a", 4097) + ".example"} {
		err := checkDomainText(domains, "content", value)
		var response *httpx.HttpError
		if !errors.As(err, &response) || response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected bad request, got %v", err)
		}
	}
}
