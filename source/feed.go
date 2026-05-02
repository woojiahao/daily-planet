package source

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/woojiahao/daily-planet/source/helpers"
)

type (
	FeedKey    string
	ArticleKey string
)

// Internal representation of an article to print.
type Article struct {
	Title      string
	Link       string
	Author     string
	Categories []string
	Id         string
	PubDate    time.Time

	// Assumed to be in markdown format, engines should implement parsing from HTML.
	Content string

	// Escape hatch into underlying feed type's data.
	EngineType helpers.EngineType
	Engine     any
}

func (article Article) GetKey() ArticleKey {
	if article.Id != "" {
		return ArticleKey(article.Id)
	}

	if article.Link != "" {
		return ArticleKey(article.Link)
	}

	if article.Title != "" {
		return ArticleKey(article.Title)
	}

	// In the worst case, use PubDate in Unix time
	return ArticleKey(fmt.Sprint(article.PubDate.Unix()))
}

// Internal representation of a feed to print.
type Feed struct {
	Title         string
	Link          string
	Description   string
	Articles      []Article
	Language      string
	Copyright     string
	PubDate       time.Time
	LastBuildDate time.Time
	Categories    []string

	// Escape hatch into underlying feed type's data.
	EngineType helpers.EngineType
	Engine     any
}

func (feed Feed) GetKey() FeedKey {
	if feed.Link != "" {
		return FeedKey(feed.Link)
	}

	if feed.Title != "" {
		return FeedKey(feed.Title)
	}

	// In the worst case, use PubDate in Unix time
	return FeedKey(fmt.Sprint(feed.PubDate.Unix()))
}

// Feeds can come in RSS, Atom, or JSON format. Support all during loading into a common object.
//
// RSS feeds contain the &lt;rss&gt; root element.
//
// Atom feeds contain the &lt;feed&gt; root element.
//
// JSON feeds are just JSON files.
func LoadFeed(feedURL string) (Feed, error) {
	// TODO(woojiahao): Log when feeds fail to load

	resp, err := http.Get(feedURL)
	if err != nil {
		return Feed{}, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Feed{}, err
	}

	// Load RSS -> Atom -> JSON in that order
	// TODO(woojiahao): Clean up this parsing logic since we might want to directly detect the file type
	rss, err := parseRSS(respBody)
	if err == nil && rss.Channel.Title != "" {
		return rss.Channel.toFeed(), nil
	}

	atom, err := parseAtom(respBody)
	if err == nil && atom.Title != "" {
		return atom.toFeed(), nil
	}

	json, err := parseJSON(respBody)
	if err == nil && json.Title != "" {
		return json.toFeed(), nil
	}

	return Feed{}, fmt.Errorf("failed to load feed, not types RSS, Atom, or JSON")
}

func BulkLoadFeeds(feedURLs []string) []Feed {
	type result struct {
		index int
		feed  Feed
	}

	workers := runtime.NumCPU() * 4
	n := len(feedURLs)

	if workers > n {
		workers = n
	}

	base := n / workers
	extra := n % workers

	grouped := make([][]string, workers)

	start := 0
	for i := range workers {
		size := base
		if i < extra {
			size++
		}

		end := start + size
		grouped[i] = feedURLs[start:end]
		start = end
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	ch := make(chan result)
	loadFeeds := func(i int) {
		defer wg.Done()
		urls := grouped[i]
		for j, url := range urls {
			feed, err := LoadFeed(url)
			// TODO(woojiahao): maintain skipped feed list
			if err == nil {
				// computing global index for the feed
				idx := 0
				for k := range i {
					idx += len(grouped[k])
				}
				idx += j
				ch <- result{index: idx, feed: feed}
			}
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for i := range workers {
		go loadFeeds(i)
	}

	feeds := make([]Feed, len(feedURLs))
	for result := range ch {
		feeds[result.index] = result.feed
	}

	return feeds
}
