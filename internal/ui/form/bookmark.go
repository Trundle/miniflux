package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strings"
)

// BookmarkForm represents a bookmark form in the UI
type BookmarkForm struct {
	URL     string
	Title   string
	Content string
	Tags    []string
}

// NewBookmarkForm returns a new BookmarkForm.
func NewBookmarkForm(r *http.Request) *BookmarkForm {
	tags := []string{}
	for _, tag := range strings.Split(r.FormValue("tags"), ",") {
		stripped := strings.TrimSpace(tag)
		if stripped != "" {
			tags = append(tags, stripped)
		}
	}

	return &BookmarkForm{
		URL:     r.FormValue("url"),
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		Tags:    tags,
	}
}
