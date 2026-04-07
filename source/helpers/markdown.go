package helpers

import (
	"html"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// ContentToMarkdown converts string content (likely HTML) into Markdown
func ContentToMarkdown(content string) string {
	s := strings.TrimSpace(content)

	// Handling CDATA markers
	if strings.HasPrefix(s, "<![CDATA[") && strings.HasSuffix(s, "]]>") {
		s = s[9 : len(s)-3]
	}

	s = html.UnescapeString(s)

	value := s
	markdownContent, err := htmltomarkdown.ConvertString(s)
	if err == nil {
		value = markdownContent
	}

	return value
}
