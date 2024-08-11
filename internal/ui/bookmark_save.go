package ui // import "miniflux.app/v2/internal/ui"

import (
	"errors"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func fetchTitle(url string, feed *model.Feed) (string, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUserAgent(feed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(feed.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(feed.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(feed.DisableHTTP2)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(url))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		return "", localizedError.Error()
	}

	htmlDocumentReader, err := charset.NewReader(
		responseHandler.Body(config.Opts.HTTPClientMaxBodySize()),
		responseHandler.ContentType(),
	)
	if err != nil {
		return "", err
	}

	doc, docErr := goquery.NewDocumentFromReader(htmlDocumentReader)
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
		title, err := fetchTitle(bookmarkForm.URL, bookmarkFeed)
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

	if _, err := h.store.RefreshFeedEntries(user.ID, bookmarkFeed.ID, entries, true); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "starred"))
}
