package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Article represents a Substack article with its content
type Article struct {
	Title       string
	Author      string
	PublishedAt time.Time
	Content     string
	URL         string
	ImageURLs   []string
}

// ScrapeSubstack extracts content from a Substack article URL
func ScrapeSubstack(url string) (*Article, error) {
	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	article := &Article{
		URL: url,
	}

	// Extract title
	article.Title = strings.TrimSpace(doc.Find("h1.post-title").Text())
	if article.Title == "" {
		article.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Extract author - try multiple selectors
	authorSelectors := []string{
		".byline-link",
		".author-name",
		".substack-author",
		".post-header .author",
		"meta[name='author']",
	}

	for _, selector := range authorSelectors {
		if selector == "meta[name='author']" {
			// Special case for meta tag
			author, exists := doc.Find(selector).Attr("content")
			if exists && author != "" {
				article.Author = strings.TrimSpace(author)
				break
			}
		} else {
			author := strings.TrimSpace(doc.Find(selector).Text())
			if author != "" {
				article.Author = author
				break
			}
		}
	}

	// If author is still empty, try to extract from URL
	if article.Author == "" {
		parts := strings.Split(url, "//")
		if len(parts) > 1 {
			domainParts := strings.Split(parts[1], ".")
			if len(domainParts) > 0 {
				article.Author = strings.Split(domainParts[0], "/")[0]
			}
		}
	}

	// Extract publish date
	dateStr := doc.Find("time").AttrOr("datetime", "")
	if dateStr != "" {
		publishDate, err := time.Parse(time.RFC3339, dateStr)
		if err == nil {
			article.PublishedAt = publishDate
		}
	}

	// Extract content
	contentSelector := ".available-content, .subscriber-content, .post-content, .body"
	contentHTML, err := doc.Find(contentSelector).Html()
	if err != nil {
		return nil, fmt.Errorf("failed to extract content: %w", err)
	}
	article.Content = contentHTML

	// Extract images
	doc.Find(contentSelector + " img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists && src != "" {
			article.ImageURLs = append(article.ImageURLs, src)
		}
	})

	return article, nil
}
