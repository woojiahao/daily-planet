package source

import (
	"encoding/xml"

	"github.com/woojiahao/daily-planet/source/helpers"
)

// Atom &lt;link&gt; specification.
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func (link AtomLink) toLink() string {
	return link.Href
}

// Atom &lt;author&gt; specification.
type AtomAuthor struct {
	Name string `xml:"name"`
}

func (author AtomAuthor) toAuthor() string {
	return author.Name
}

// Atom &lt;category&gt; specification.
type AtomCategory struct {
	Term string `xml:"term,attr"`
}

func (category AtomCategory) toCategory() string {
	return category.Term
}

// Atom &lt;content&gt; specification.
type AtomContent struct {
	Type string `xml:"type,attr"`
	Body string `xml:",innerxml"`
}

func (content AtomContent) toContent() string {
	if content.Type == "html" || content.Type == "text" {
		return helpers.ContentToMarkdown(content.Body)
	}

	// Without a good parser, just return the raw data
	return content.Body
}

// Atom &lt;entry&gt; specification.
type AtomEntry struct {
	Title      string         `xml:"title"`
	Id         string         `xml:"id"`
	Summary    string         `xml:"summary"`
	Content    AtomContent    `xml:"content"`
	Updated    string         `xml:"updated"`
	Links      []AtomLink     `xml:"link"`
	Author     AtomAuthor     `xml:"author"`
	Categories []AtomCategory `xml:"category"`
}

func (entry AtomEntry) toArticle() Article {
	var articleLink string
	for _, link := range entry.Links {
		if link.Rel == "" || link.Rel == "alternate" {
			articleLink = link.toLink()
			break
		}
	}

	var categories []string
	for _, category := range entry.Categories {
		categories = append(categories, category.toCategory())
	}

	pubDate := helpers.DateStringToRFC3339Time(entry.Updated)

	return Article{
		Title:      entry.Title,
		Link:       articleLink,
		Author:     entry.Author.toAuthor(),
		Categories: categories,
		Id:         entry.Id,
		PubDate:    pubDate,
		Content:    entry.Content.toContent(),
		EngineType: helpers.EngineTypeAtom,
		Engine:     entry,
	}
}

// Atom &lt;feed&gt; specification.
type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`

	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle"`
	Id       string      `xml:"id"`
	Updated  string      `xml:"updated"`
	Links    []AtomLink  `xml:"link"`
	Entries  []AtomEntry `xml:"entry"`
	Author   AtomAuthor  `xml:"author"`
}

func (feed AtomFeed) toFeed() Feed {
	var feedLink string
	for _, link := range feed.Links {
		if link.Rel == "" || link.Rel == "alternate" {
			feedLink = link.toLink()
			break
		}
	}

	pubDate := helpers.DateStringToRFC3339Time(feed.Updated)

	var articles []Article
	for _, entry := range feed.Entries {
		articles = append(articles, entry.toArticle())
	}

	return Feed{
		Title:         feed.Title,
		Link:          feedLink,
		Description:   feed.Subtitle,
		Articles:      articles,
		Language:      "",
		Copyright:     "",
		PubDate:       pubDate,
		LastBuildDate: pubDate,
		Categories:    []string{},
		EngineType:    helpers.EngineTypeAtom,
		Engine:        feed,
	}
}

func parseAtom(body []byte) (AtomFeed, error) {
	var atomRaw AtomFeed
	err := xml.Unmarshal(body, &atomRaw)
	if err != nil {
		return AtomFeed{}, err
	}

	return atomRaw, nil
}
