package subject

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func checkerFor(t *testing.T, base string) *HTTPChecker {
	t.Helper()
	c := NewHTTPChecker()
	resource := registry["novel"]
	resource.resolve = func(key string) (string, error) {
		endpoint, err := novelResolver(key)
		if err != nil {
			return "", err
		}
		return base + strings.TrimPrefix(endpoint, "https://n.novelia.cc"), nil
	}
	c.registry = map[string]resourceType{"novel": resource}
	return c
}

func TestCheckStatus(t *testing.T) {
	for _, status := range []int{204, 400, 404, 200, 401, 403, 429, 500, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			}))
			defer server.Close()
			c := checkerFor(t, server.URL)
			id, err := c.Check(context.Background(), "novel", "web-syosetu-n1234")
			switch status {
			case 204:
				if id != 1 {
					t.Fatalf("type ID=%d", id)
				}
				if err != nil {
					t.Fatal(err)
				}
			case 400:
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			case 404:
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("got %v", err)
				}
			default:
				if err == nil || errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			}
		})
	}
}

func TestCheckRoutesAndEscapesKey(t *testing.T) {
	for _, tc := range []struct{ key, path string }{
		{"web-syosetu-n1234-part-2", "/api/novel/syosetu/n1234-part-2/exist"},
		{"wenku-507f1f77bcf86cd799439011", "/api/wenku/507f1f77bcf86cd799439011/exist"},
		{"web-provider-a/b?c#d%", "/api/novel/provider/a%2Fb%3Fc%23d%25/exist"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.RawQuery != "" || r.URL.EscapedPath() != tc.path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			if _, err := checkerFor(t, server.URL).Check(context.Background(), "novel", tc.key); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInvalidNovelKey(t *testing.T) {
	c := checkerFor(t, "http://invalid.invalid")
	for _, key := range []string{"", "unknown-id", "web", "web--id", "web-provider-", "wenku-"} {
		if _, err := c.Check(context.Background(), "novel", key); !errors.Is(err, ErrInvalid) {
			t.Errorf("key=%q got %v", key, err)
		}
	}
}

func TestCheckDoesNotFollowRedirect(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}))
	defer server.Close()
	c := checkerFor(t, server.URL)
	if _, err := c.Check(context.Background(), "novel", "web-syosetu-n1234"); err == nil {
		t.Fatal("redirect accepted")
	}
	if calls != 1 {
		t.Fatalf("followed redirect: %d requests", calls)
	}
}

func TestCheckTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	c := checkerFor(t, server.URL)
	c.client.Timeout = 20 * time.Millisecond
	if _, err := c.Check(context.Background(), "novel", "web-syosetu-n1234"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Check(ctx, "novel", "web-syosetu-n1234"); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestSupportedKinds(t *testing.T) {
	c := NewHTTPChecker()
	if len(c.registry) != 1 {
		t.Fatalf("unexpected supported kinds: %v", c.registry)
	}
	endpoint, err := c.registry["novel"].resolve("web-syosetu-n1234")
	if err != nil || endpoint != "https://n.novelia.cc/api/novel/syosetu/n1234/exist" {
		t.Fatalf("endpoint=%s error=%v", endpoint, err)
	}
	if _, err := c.Check(context.Background(), "unknown", "key"); err == nil {
		t.Fatal("unsupported kind accepted")
	}
}

func TestTypeID(t *testing.T) {
	if id, ok := TypeID("novel"); !ok || id != 1 {
		t.Fatalf("id=%d ok=%v", id, ok)
	}
	if _, ok := TypeID("unknown"); ok {
		t.Fatal("unknown kind accepted")
	}
}
