package helpers

import (
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/ds"
	"github.com/woojiahao/daily-planet/source"
)

func FetchNewArticles(feed source.Feed, cachedArticles []models.Cache) ([]source.Article, []string) {
	cachedArticleKeys := ds.NewSet[source.ArticleKey]()
	for _, article := range cachedArticles {
		cachedArticleKeys.Add(source.ArticleKey(article.ArticleKey))
	}

	var newArticles []source.Article
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
