// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showTagsPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

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

	tags, err := h.store.AllTags(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	view.Set("tags", tags)
	view.Set("entry", entry)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	html.OK(w, r, view.Render("show_tags"))
}
