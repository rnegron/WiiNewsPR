package endi

import (
	"WiiNewsPR/news"
	_ "embed"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	SanJuanLatitude  = 18.466333
	SanJuanLongitude = -66.105721
)

const MaxArticles = 15
const MaxArticlesPerCategory = 3

//go:embed logo.jpg
var logo []byte

func (e *Endi) GetLogo() []byte {
	return logo
}

// RSS feed structures
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title        string       `xml:"title"`
	Description  string       `xml:"description"`
	MediaContent MediaContent `xml:"http://search.yahoo.com/mrss/ content"`
}

type MediaContent struct {
	URL         string `xml:"url,attr"`
	Type        string `xml:"type,attr"`
	Description string `xml:"http://search.yahoo.com/mrss/ description"`
}

func (e *Endi) makeRequest(client *http.Client, feedURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "WiiNewsPR/1.0 (+https://github.com/rnegron/WiiNewsPR)")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("RSS feed returned status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func (e *Endi) GetArticles() ([]news.Article, error) {

	baseURL := "https://www.elnuevodia.com/arc/outboundfeeds/rss/category/%s/?outputType=xml"

	feeds := map[news.Topic]string{
		news.NationalNews:  fmt.Sprintf(baseURL, "noticias/locales"),
		news.Sports:        fmt.Sprintf(baseURL, "deportes"),
		news.Entertainment: fmt.Sprintf(baseURL, "entretenimiento"),
		news.Business:      fmt.Sprintf(baseURL, "negocios"),
		news.Science:       fmt.Sprintf(baseURL, "ciencia-ambiente"),
		news.Technology:    fmt.Sprintf(baseURL, "tecnologia"),
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	allArticles := []news.Article{}

	for topic, feedURL := range feeds {
		articles, err := e.fetchFromFeed(client, feedURL, topic)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch feed: %v\n", err)
			continue
		}
		allArticles = append(allArticles, articles...)

	}

	if len(allArticles) > MaxArticles {
		allArticles = allArticles[:MaxArticles]
	}

	return allArticles, nil
}
func (e *Endi) fetchFromFeed(client *http.Client, feedURL string, topic news.Topic) ([]news.Article, error) {
	resp, err := e.makeRequest(client, feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS response: %v", err)
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %v", err)
	}

	if err != nil {
		return nil, err
	}

	return e.convertRSSToArticles(&rss, topic), nil
}

func (e *Endi) convertRSSToArticles(rss *RSS, topic news.Topic) []news.Article {
	articles := []news.Article{}

	for i, item := range rss.Channel.Items {
		if i >= MaxArticlesPerCategory {
			break
		}

		title := strings.TrimSpace(item.Title)
		if isDuplicate(title, e.oldArticleTitles) {
			continue
		}

		article := e.createArticleFromItem(item, topic, title)
		articles = append(articles, article)
	}

	return articles
}

func (e *Endi) createArticleFromItem(item Item, topic news.Topic, title string) news.Article {
	article := news.Article{
		Title: title,
		Topic: topic,
		Location: &news.Location{
			Name:      "San Juan",
			Latitude:  SanJuanLatitude,
			Longitude: SanJuanLongitude,
		},
	}

	if item.Description != "" {
		content := cleanDescription(item.Description)
		article.Content = &content
	}

	if item.MediaContent.URL != "" && item.MediaContent.Type == "image/jpeg" {
		thumbnail := e.createThumbnail(item.MediaContent, title)
		if thumbnail != nil {
			article.Thumbnail = thumbnail
		}
	}

	return article
}

func (e *Endi) createThumbnail(mediaContent MediaContent, fallbackCaption string) *news.Thumbnail {
	imageData, err := news.DownloadImage(mediaContent.URL)
	if err != nil || len(imageData) == 0 {
		return nil
	}

	convertedImage := news.ConvertImage(imageData)
	if len(convertedImage) == 0 {
		return nil
	}

	caption := mediaContent.Description
	if caption == "" {
		caption = fallbackCaption
	}

	return &news.Thumbnail{
		Image:   convertedImage,
		Caption: cleanDescription(caption),
	}
}

func isDuplicate(title string, oldTitles []string) bool {
	for _, oldTitle := range oldTitles {
		if strings.EqualFold(title, oldTitle) {
			return true
		}
	}
	return false
}

func cleanDescription(description string) string {
	description = html.UnescapeString(description)

	// Basic HTML tag removal
	replacements := map[string]string{
		"<p>":    "",
		"</p>":   "",
		"<br>":   " ",
		"<br/>":  " ",
		"<br />": " ",
	}

	for old, new := range replacements {
		description = strings.ReplaceAll(description, old, new)
	}

	return strings.TrimSpace(description)
}
