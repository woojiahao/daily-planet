package source

import (
	"encoding/xml"

	"github.com/woojiahao/daily-planet/source/helpers"
)

type RSSContent struct {
	Body string `xml:",innerxml"`
}

// RSS &lt;item&gt; specification.
//
// Source: https://www.rssboard.org/rss-specification#hrelementsOfLtitemgt
type RSSItem struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description RSSContent `xml:"description"`
	Author      string     `xml:"author"`
	Categories  []string   `xml:"category"`
	Guid        string     `xml:"guid"`
	PubDate     string     `xml:"pubDate"`
}

// Converts an RSSItem into a Feed.
func (item RSSItem) toArticle() Article {
	pubDate := helpers.DateStringToRFC822Time(item.PubDate)
	content := helpers.ContentToMarkdown(item.Description.Body)

	return Article{
		Title:      item.Title,
		Link:       item.Link,
		Content:    content,
		Author:     item.Author,
		Categories: item.Categories,
		ID:         item.Guid,
		PubDate:    pubDate,
		EngineType: helpers.EngineTypeRSS,
		Engine:     item,
	}
}

// TODO(woojiahao): Do something with the RSSImage

// RSS &lt;image&gt; specification.
//
// Source: https://www.rssboard.org/rss-specification#ltimagegtSubelementOfLtchannelgt
type RSSImage struct {
	URL    string `xml:"url"`
	Title  string `xml:"title"`
	Width  uint8  `xml:"width"`
	Height uint8  `xml:"height"`
}

// TODO(woojiahao): Implement skipHours and skipDays to abide closely with specification

// RSS &lt;channel&gt; specification.
//
// Source: https://www.rssboard.org/rss-specification#hrelementsOfLtitemgt
type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Items         []RSSItem `xml:"item"`
	Language      string    `xml:"language"`
	Copyright     string    `xml:"copyright"`
	PubDate       string    `xml:"pubDate"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Categories    []string  `xml:"category"`
	Image         RSSImage  `xml:"image"`
}

// RSS &lt;rss&gt; root element specification.
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Channel RSSChannel `xml:"channel"`
}

// Converts an RSSChannel into a Feed.
func (channel RSSChannel) toFeed() Feed {
	pubDate := helpers.DateStringToRFC822Time(channel.PubDate)
	lastBuildDate := helpers.DateStringToRFC822Time(channel.LastBuildDate)

	var articles []Article
	for _, item := range channel.Items {
		articles = append(articles, item.toArticle())
	}

	return Feed{
		Title:         channel.Title,
		Link:          channel.Link,
		Description:   channel.Description,
		Articles:      articles,
		Language:      channel.Language,
		Copyright:     channel.Copyright,
		PubDate:       pubDate,
		LastBuildDate: lastBuildDate,
		Categories:    channel.Categories,
		EngineType:    helpers.EngineTypeRSS,
		Engine:        channel,
	}
}

func parseRSS(body []byte) (RSS, error) {
	var rssRaw RSS
	err := xml.Unmarshal(body, &rssRaw)
	if err != nil {
		return RSS{}, err
	}

	return rssRaw, nil
}
