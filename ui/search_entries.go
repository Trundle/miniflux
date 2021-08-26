// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"
	"strings"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func parseQuery(q string) (string, []string) {
	tags := []string{}
	words := []string{}
	for _, word := range strings.Split(q, " ") {
		if strings.HasPrefix(word, "tag:") {
			tags = append(tags, word[len("tag:"):])
		} else {
			words = append(words, word)
		}
	}
	return strings.Join(words, " "), tags
}

func buildQuery(q string, tags []string) string {
	result := ""
	if len(tags) > 0 {
		for _, tag := range tags {
			result += "tag:" + tag + " "
		}
	}
	result += q
	return result
}

func (h *handler) showSearchEntriesPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	searchQuery, searchTags := parseQuery(request.QueryStringParam(r, "q", ""))
	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithSearchQuery(searchQuery)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)
	if len(searchTags) > 0 {
		builder.WithTags(searchTags)
	}

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	pagination := getPagination(route.Path(h.router, "searchEntries"), count, offset, user.EntriesPerPage)
	pagination.SearchQuery = searchQuery

	view.Set("searchQuery", buildQuery(searchQuery, searchTags))
	view.Set("entries", entries)
	view.Set("total", count)
	view.Set("pagination", pagination)
	view.Set("menu", "search")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("search_entries"))
}
