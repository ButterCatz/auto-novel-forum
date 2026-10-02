package subject

import (
	"net/url"
	"strings"
)

// Key formats: web-{providerId}-{novelId} or wenku-{novelId}.
// Web novel IDs may contain hyphens.
func novelResolver(key string) (string, error) {
	const base = "https://n.novelia.cc/api"
	if id, ok := strings.CutPrefix(key, "wenku-"); ok && id != "" {
		return base + "/wenku/" + url.PathEscape(id) + "/exist", nil
	}
	parts := strings.SplitN(key, "-", 3)
	if len(parts) == 3 && parts[0] == "web" && parts[1] != "" && parts[2] != "" {
		return base + "/novel/" + url.PathEscape(parts[1]) + "/" + url.PathEscape(parts[2]) + "/exist", nil
	}
	return "", ErrInvalid
}
