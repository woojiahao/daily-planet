package source

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/source/helpers"
)

// JSON feed v1.1 attachment element
//
// Based on https://www.jsonfeed.org/version/1.1/#attachments
type JSONAttachment struct {
	URL               string `json:"url"`
	MimeType          string `json:"mime_type"`
	Title             string `json:"title"`
	SizeInBytes       uint8  `json:"size_in_bytes"`
	DurationInSeconds uint8  `json:"duration_in_seconds"`
}

// JSON feed v1.1 author element
//
// Based on https://www.jsonfeed.org/version/1.1/#top-level
type JSONAuthor struct {
	Name   string `json:"author"`
	URL    string `json:"url"`
	Avatar string `json:"avatar"`
}

func (author JSONAuthor) toAuthor() string {
	return author.Name
}

// JSON feed v1.1 item element
//
// Based on https://www.jsonfeed.org/version/1.1/#items
type JSONItem struct {
	Id            string           `json:"id"`
	URL           string           `json:"url"`
	ExternalURL   string           `json:"external_url"`
	Title         string           `json:"title"`
	ContentHTML   string           `json:"content_html"`
	ContentText   string           `json:"content_text"`
	Summary       string           `json:"summary"`
	Image         string           `json:"image"`
	BannerImage   string           `json:"banner_image"`
	DatePublished string           `json:"date_published"`
	DateModified  string           `json:"date_modified"`
	Author        JSONAuthor       `json:"author"`
	Authors       []JSONAuthor     `json:"authors"`
	Tags          []string         `json:"tags"`
	Language      string           `json:"language"`
	Attachments   []JSONAttachment `json:"attachments"`
}

func (item JSONItem) toArticle() Article {
	var author string
	if len(item.Authors) == 0 {
		author = item.Author.toAuthor()
	} else {
		var authorNames []string
		for _, a := range item.Authors {
			authorNames = append(authorNames, a.toAuthor())
		}
		author = strings.Join(authorNames, ", ")
	}

	pubDate := helpers.DateStringToRFC3339Time(item.DatePublished)
	content := item.ContentText
	if item.ContentHTML != "" {
		content = helpers.ContentToMarkdown(item.ContentHTML)
	}

	return Article{
		Title:      item.Title,
		Link:       item.URL,
		Author:     author,
		Categories: item.Tags,
		Id:         item.Id,
		PubDate:    pubDate,
		Content:    content,
		EngineType: helpers.EngineTypeJSON,
		Engine:     item,
	}
}

// JSON feed v1.1 hub element
//
// Based on https://www.jsonfeed.org/version/1.1/#top-level
type JSONHub struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// JSON feed v1.1 feed root element
//
// Based on https://www.jsonfeed.org/version/1.1/#top-level
type JSONFeed struct {
	Version     string       `json:"version"`
	Title       string       `json:"title"`
	HomePageURL string       `json:"home_page_url"`
	FeedURL     string       `json:"feed_url"`
	Description string       `json:"description"`
	UserComment string       `json:"user_comment"`
	NextURL     string       `json:"next_url"`
	Icon        string       `json:"icon"`
	Favicon     string       `json:"favicon"`
	Author      JSONAuthor   `json:"author"`
	Authors     []JSONAuthor `json:"authors"`
	Language    string       `json:"language"`
	Expired     bool         `json:"expired"`
	Hubs        []JSONHub    `json:"hubs"`
	Items       []JSONItem   `json:"items"`
}

func (feed JSONFeed) toFeed() Feed {
	var articles []Article
	for _, a := range feed.Items {
		articles = append(articles, a.toArticle())
	}

	return Feed{
		Title:         feed.Title,
		Link:          feed.FeedURL,
		Description:   feed.Description,
		Articles:      articles,
		Language:      feed.Language,
		Copyright:     "",
		PubDate:       time.Time{},
		LastBuildDate: time.Time{},
		Categories:    []string{},
		EngineType:    helpers.EngineTypeJSON,
		Engine:        feed,
	}
}

func parseJSON(body []byte) (JSONFeed, error) {
	var jsonRaw JSONFeed
	err := json.Unmarshal(body, &jsonRaw)
	if err != nil {
		return JSONFeed{}, err
	}

	return jsonRaw, nil
}
