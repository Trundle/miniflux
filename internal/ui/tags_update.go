package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) updateTags(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entry == nil {
		html.NotFound(w, r)
		return
	}

	tagsForm := form.NewTagsForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", tagsForm)
	view.Set("entry", entry)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	entry.Tags = tagsForm.Tags;

	if err := h.store.UpdateTags(entry); err != nil {
		slog.Error("[UI:UpdateTags] Unable to update tags",
			slog.Any("error", err),
		)
		view.Set("errorMessage", "Unabe to update tags.")
		html.OK(w, r, view.Render("edit_tags"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "starredEntry", "entryID", entryID))
}
