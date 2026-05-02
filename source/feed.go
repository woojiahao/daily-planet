package source

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/ds"
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

func FetchNewArticles(feed Feed, cachedArticles []models.Cache) ([]Article, []string) {
	cachedArticleKeys := ds.NewSet[ArticleKey]()
	for _, article := range cachedArticles {
		cachedArticleKeys.Add(ArticleKey(article.ArticleKey))
	}

	var newArticles []Article
	var newArticleKeys []string
	for _, article := range feed.Articles {
		if !cachedArticleKeys.Contains(article.GetKey()) {
			// new article to cache and print
			newArticles = append(newArticles, article)
			newArticleKeys = append(newArticleKeys, string(article.GetKey()))
		}
	}

	return newArticles, newArticleKeys
}

func FetchFeedsAlgorithmWrapper(configurationID models.ConfigurationID, database *db.Database, sender func(title, description string, color common.Color)) {
	// Fetching algorithm:
	// 1. Fetch all enabled feeds for this configuration
	// 2. Map the feed ID -> feed / feed URL from the enabled feed
	// 3. Retrieve all of the caches for the enabled feeds
	// 4. Fetch all feeds from all sources concurrently
	// 5. For each feed retrieved,
	// 	a. For each article retrieved,
	// 		1) Check if article exists within cache
	// 		2) If article exists within cache, skip over it
	// 		3) If article does not exist within cache, add to be inserted and add to new articles

	enabledFeeds, err := database.Feed.AllEnabledByConfigurationID(configurationID)
	if err != nil {
		fmt.Printf("err is %v\n", err)
		sender(
			"Failed to fetch feeds",
			"Failed to fetch feeds for this source.",
			common.ColorRed,
		)
		return
	}

	fmt.Println("enabled feeds:")
	for _, feed := range enabledFeeds {
		fmt.Printf("- %s\n", feed.URL)
	}

	feedURLs := make(map[models.FeedID]string)
	feedMap := make(map[models.FeedID]models.Feed)
	var cacheKeys []models.CacheKey
	for _, feed := range enabledFeeds {
		feedMap[feed.ID] = feed
		feedURLs[feed.ID] = feed.URL
		cacheKeys = append(cacheKeys, models.NewCacheKey(feed.ConfigurationID, feed.ID))
	}
	caches, err := database.Cache.AllByKeys(cacheKeys)
	if err != nil {
		fmt.Printf("err is %v\n", err)
		sender(
			"Failed to fetch feeds",
			"Failed to fetch cache for feeds in this source.",
			common.ColorRed,
		)
		return
	}

	feedCaches := make(map[models.FeedID][]models.Cache)
	for _, cache := range caches {
		feedCaches[cache.FeedID] = append(feedCaches[cache.FeedID], cache)
	}

	fmt.Println("feed caches:")
	for k, v := range feedCaches {
		fmt.Printf("cache for %d with %d entries\n", k, len(v))
	}

	urls := slices.Sorted(maps.Values(feedURLs))
	feeds := BulkLoadFeeds(urls)
	feedsByURLs := make(map[string]Feed)
	for i, url := range urls {
		feedsByURLs[url] = feeds[i]
	}

	var insertCacheKeys []models.CacheKey
	var insertArticleKeys []string
	var allArticles []Article
	for feedID := range feedURLs {
		feedCache := feedCaches[feedID]

		newArticles, newArticleKeys := FetchNewArticles(
			feedsByURLs[feedURLs[feedID]],
			feedCache,
		)

		allArticles = append(allArticles, newArticles...)
		for _, articleKey := range newArticleKeys {
			cacheKey := models.NewCacheKey(configurationID, feedID)
			insertCacheKeys = append(insertCacheKeys, cacheKey)
			insertArticleKeys = append(insertArticleKeys, articleKey)
		}
	}

	if len(allArticles) != 0 {
		err = database.Cache.InsertMany(insertCacheKeys, insertArticleKeys)
		if err != nil {
			fmt.Printf("%v\n", err)
			sender(
				"Failed to save cache",
				"Failed to save articles into cache in this source",
				common.ColorRed,
			)
			return
		}

		var articleStrings []string
		for _, article := range allArticles {
			articleStrings = append(articleStrings, fmt.Sprintf("- [%s](%s)", article.Title, article.Link))
		}

		fmt.Printf("articleStrings total length %d\n", len(articleStrings))

		var groupedArticleStrings [][]string
		groupedArticleStrings = append(groupedArticleStrings, []string{})
		acc := 0
		const limit = 3500
		for _, articleString := range articleStrings {
			// include \n at the end as well
			if acc+len(articleString)+1 > limit {
				// split out
				groupedArticleStrings = append(groupedArticleStrings, []string{})
				acc = len(articleString) + 1
			} else {
				acc += len(articleString) + 1
			}
			groupedArticleStrings[len(groupedArticleStrings)-1] = append(groupedArticleStrings[len(groupedArticleStrings)-1], articleString)
		}

		fmt.Printf("gropued article strings length %d\n", len(groupedArticleStrings))

		if len(groupedArticleStrings) > 0 {
			for _, group := range groupedArticleStrings {
				sender(
					"Feeds fetched",
					strings.Join(group, "\n"),
					common.ColorBlue,
				)
			}
		}

		return
	}

	sender(
		"Feeds fetched",
		"All feeds fetched but no new articles found",
		common.ColorBlue,
	)
}
