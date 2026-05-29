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

	"github.com/woojiahao/daily-planet/apperrors"
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

// Article is the internal representation of an article to print.
type Article struct {
	Title      string
	Link       string
	Author     string
	Categories []string
	ID         string
	PubDate    time.Time

	// Assumed to be in markdown format, engines should implement parsing from HTML.
	Content string

	// Escape hatch into underlying feed type's data.
	EngineType helpers.EngineType
	Engine     any
}

func (article Article) GetKey() ArticleKey {
	if article.ID != "" {
		return ArticleKey(article.ID)
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

// Feed is the internal representation of a feed to print.
type Feed struct {
	Title         string
	Link          string
	RawLink       string
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

	client := &http.Client{}
	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return Feed{}, common.WrapError(apperrors.ErrLoadFeedFailed, err)
	}

	req.Header.Set("User-Agent", "daily-planet/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return Feed{}, common.WrapError(apperrors.ErrLoadFeedFailed, err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Feed{}, common.WrapError(apperrors.ErrLoadFeedFailed, err)
	}

	var feed Feed

	// Load RSS -> Atom -> JSON in that order
	// TODO(woojiahao): Clean up this parsing logic since we might want to directly detect the file type
	rss, err := parseRSS(respBody)
	if err == nil && rss.Channel.Title != "" {
		feed = rss.Channel.toFeed()
	} else {
		atom, err := parseAtom(respBody)
		if err == nil && atom.Title != "" {
			feed = atom.toFeed()
		} else {
			json, err := parseJSON(respBody)
			if err == nil && json.Title != "" {
				feed = json.toFeed()
			}
		}
	}

	if feed.Title != "" {
		feed.RawLink = feedURL
		return feed, nil
	}

	return Feed{}, common.WrapError(apperrors.ErrLoadFeedFailed, apperrors.ErrLoadFeedUnsuppportedType)
}

func BulkLoadFeeds(feedURLs []string) []Feed {
	if len(feedURLs) == 0 {
		// no feeds, skip the entire process
		return []Feed{}
	}

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

func FetchFeedAlgorithmWrapper(
	key models.FeedKey,
	database *db.Database,
	sendNoArticlesUpdate bool,
	sender func(title, description string, color common.Color),
) {
	feed, err := database.Feed.OneByKey(key)
	if err != nil {
		fmt.Printf("err is %v\n", err)

		sender(
			"Failed to fetch feed",
			"Failed to fetch feed for this source.",
			common.ColorRed,
		)

		return
	}

	fetchFeeds(
		[]models.Feed{feed},
		database,
		sendNoArticlesUpdate,
		sender,
	)
}

func FetchFeedsAlgorithmWrapper(
	configurationID models.ConfigurationID,
	database *db.Database,
	sendNoArticlesUpdate bool,
	sender func(title, description string, color common.Color),
) {
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

	if len(enabledFeeds) == 0 {
		if sendNoArticlesUpdate {
			sender(
				"No feeds",
				"Add feeds first before trying to fetch all",
				common.ColorRed,
			)
		}

		return
	}

	fetchFeeds(
		enabledFeeds,
		database,
		sendNoArticlesUpdate,
		sender,
	)
}

type feedArticle struct {
	feed    Feed
	article Article
}

func fetchFeeds(
	feeds []models.Feed,
	database *db.Database,
	sendNoArticlesUpdate bool,
	sender func(title, description string, color common.Color),
) {
	fmt.Println("enabled feeds:")

	for _, feed := range feeds {
		fmt.Printf("- %s\n", feed.URL)
	}

	feedURLs := make(map[models.FeedID]string)

	var cacheKeys []models.CacheKey

	for _, feed := range feeds {
		feedURLs[feed.ID] = feed.URL

		cacheKeys = append(
			cacheKeys,
			models.NewCacheKey(feed.ConfigurationID, feed.ID),
		)
	}

	caches, err := database.Cache.All(cacheKeys...)
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
		feedCaches[cache.FeedID] = append(
			feedCaches[cache.FeedID],
			cache,
		)
	}

	fmt.Println("feed caches:")

	for k, v := range feedCaches {
		fmt.Printf(
			"cache for %d with %d entries\n",
			k,
			len(v),
		)
	}

	urls := slices.Sorted(maps.Values(feedURLs))

	var loadedFeeds []Feed

	if len(urls) == 1 {
		feed, err := LoadFeed(urls[0])
		if err != nil {
			fmt.Printf("err is %v\n", err)

			sender(
				"Failed to fetch feeds",
				"Failed to load articles for feed in this source.",
				common.ColorRed,
			)

			return
		}

		loadedFeeds = []Feed{feed}
	} else {
		loadedFeeds = BulkLoadFeeds(urls)
	}

	feedsByURL := make(map[string]Feed)

	for i, url := range urls {
		feedsByURL[url] = loadedFeeds[i]
	}

	var (
		cacheInserts []models.CacheInsert
		allArticles  []feedArticle
	)

	for _, feed := range feeds {
		newArticles, newArticleKeys := FetchNewArticles(
			feedsByURL[feed.URL],
			feedCaches[feed.ID],
		)

		for _, article := range newArticles {
			allArticles = append(allArticles, feedArticle{
				feed:    feedsByURL[feed.URL],
				article: article,
			})
		}

		for _, articleKey := range newArticleKeys {
			cacheInserts = append(
				cacheInserts,
				models.CacheInsert{
					CacheKey: models.NewCacheKey(
						feed.ConfigurationID,
						feed.ID,
					),
					ArticleKey: articleKey,
				},
			)
		}
	}

	if len(allArticles) == 0 {
		if sendNoArticlesUpdate {
			sender(
				"Feeds fetched",
				"All feeds fetched but no new articles found",
				common.ColorBlue,
			)
		}

		return
	}

	err = database.Cache.Insert(cacheInserts...)
	if err != nil {
		fmt.Printf("%v\n", err)

		sender(
			"Failed to save cache",
			"Failed to save articles into cache in this source",
			common.ColorRed,
		)

		return
	}

	sendArticles(allArticles, sender)
}

func sendArticles(
	feedArticles []feedArticle,
	sender func(title, description string, color common.Color),
) {
	var articleStrings []string

	for _, feedArticle := range feedArticles {
		articleStrings = append(
			articleStrings,
			fmt.Sprintf(
				"- [%s] [%s](%s)",
				feedArticle.feed.Title,
				feedArticle.article.Title,
				feedArticle.article.Link,
			),
		)
	}

	fmt.Printf(
		"articleStrings total length %d\n",
		len(articleStrings),
	)

	const limit = 3500

	var groupedArticleStrings [][]string

	groupedArticleStrings = append(
		groupedArticleStrings,
		[]string{},
	)

	acc := 0

	for _, articleString := range articleStrings {
		size := len(articleString) + 1

		if acc+size > limit {
			groupedArticleStrings = append(
				groupedArticleStrings,
				[]string{},
			)

			acc = size
		} else {
			acc += size
		}

		groupedArticleStrings[len(groupedArticleStrings)-1] = append(
			groupedArticleStrings[len(groupedArticleStrings)-1],
			articleString,
		)
	}

	fmt.Printf(
		"grouped article strings length %d\n",
		len(groupedArticleStrings),
	)

	for _, group := range groupedArticleStrings {
		sender(
			"Feeds fetched",
			strings.Join(group, "\n"),
			common.ColorBlue,
		)
	}
}
