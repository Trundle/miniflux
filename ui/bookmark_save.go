// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"errors"
	"net/http"
	"time"

	"miniflux.app/config"
	"miniflux.app/crypto"
	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/reader/browser"
	"miniflux.app/ui/form"

	"github.com/PuerkitoBio/goquery"
)

func fetchTitle(url string, userAgent string) (string, error) {
	clt := client.NewClientWithConfig(url, config.Opts)
	if len(userAgent) > 0 {
		clt.WithUserAgent(userAgent)
	}
	response, err := browser.Exec(clt)
	if err != nil {
		return "", err
	}

	doc, docErr := goquery.NewDocumentFromReader(response.Body)
	if docErr != nil {
		return "", docErr
	}

	return doc.Find("title").First().Text(), nil
}

func (h *handler) saveStarredEntry(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	bookmarkForm := form.NewBookmarkForm(r)

	feeds, err := h.store.Feeds(user.ID);
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	var bookmarkFeed *model.Feed
	for _, feed := range feeds {
		if feed.Category.Title == "Bookmarks" {
			bookmarkFeed = feed
			break
		}
	}
	if bookmarkFeed == nil {
		html.ServerError(w, r, errors.New("No bookmark feed found"))
		return
	}

	if len(bookmarkForm.Title) == 0 {
		title, err := fetchTitle(bookmarkForm.URL, bookmarkFeed.UserAgent)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		bookmarkForm.Title = title
	}

	entry := model.Entry{
		URL: bookmarkForm.URL,
		Title: bookmarkForm.Title,
		Content: bookmarkForm.Content,
		Tags: bookmarkForm.Tags,
		Date: time.Now(),
		CreatedAt: time.Now(),
		Hash: crypto.Hash(bookmarkForm.URL),
		Starred: true,
	}
	entries := make(model.Entries, 0)
	entries = append(entries, &entry)

	if err := h.store.RefreshFeedEntries(user.ID, bookmarkFeed.ID, entries, true); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "starred"))
}
