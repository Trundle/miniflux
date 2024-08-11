package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strings"
)

// TagsForm represents a tags of some entry form in the UI
type TagsForm struct {
	Tags        []string
}

// NewTagsForm returns a new TagsForm.
func NewTagsForm(r *http.Request) *TagsForm {
	tags := []string{}
	for _, tag := range strings.Split(r.FormValue("tags"), ",") {
		stripped := strings.TrimSpace(tag)
		if stripped != "" {
			tags = append(tags, stripped)
		}
	}

	return &TagsForm{
		Tags: tags,
	}
}
