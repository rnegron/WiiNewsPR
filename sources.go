package main

import (
	"WiiNewsPR/news/endi"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"
)


type Source struct {
	Logo            uint8
	Position        uint8
	_               uint16
	PictureSize     uint32
	PictureOffset   uint32
	NameSize        uint32
	NameOffset      uint32
	CopyrightSize   uint32
	CopyrightOffset uint32
}


func (n *News) GetNewsArticles() {
	n.newsSource = endi.NewEndi(n.oldArticleTitles)
	var err error
	n.articles, err = n.newsSource.GetArticles()
	if err != nil {
		panic(err)
	}

	// Save articles to file for inspection (Debug)
	n.debugSaveArticles()
}

func (n *News) MakeSourceTable() {
	n.Header.SourceTableOffset = n.GetCurrentSize()

	logo := n.newsSource.GetLogo()
	n.Sources = append(n.Sources, Source{
		Logo:            0,
		Position:        1,
		PictureSize:     uint32(len(logo)),
		PictureOffset:   0,
		NameSize:        0,
		NameOffset:      0,
		CopyrightSize:   0,
		CopyrightOffset: 0,
	})

	n.Sources[0].PictureOffset = n.GetCurrentSize()
	n.SourcePictures = logo

	for n.GetCurrentSize()%4 != 0 {
		n.SourcePictures = append(n.SourcePictures, 0)
	}

	n.Header.NumberOfSources = 1
}

// debugSaveArticles saves the fetched articles to a readable JSON file so you can see what was fetched.
func (n *News) debugSaveArticles() {
	if len(n.articles) == 0 {
		fmt.Printf("No articles found\n")
		return
	}

	// Create directory
	err := os.MkdirAll("debug", 0755)
	if err != nil {
		fmt.Printf("Error creating debug directory: %v\n", err)
		return
	}

	// Structure
	type DebugArticle struct {
		Title        string `json:"title"`
		Content      string `json:"content"`
		Topic        string `json:"topic"`
		Location     string `json:"location"`
		HasImage     bool   `json:"hasImage"`
		ImageSize    int    `json:"imageSize"`
		ImageCaption string `json:"imageCaption"`
	}

	var debugArticles []DebugArticle
	topicNames := []string{"National", "International", "Sports", "Entertainment", "Business", "Science", "Technology"}

	for _, article := range n.articles {
		var content string
		if article.Content != nil {
			content = *article.Content
		} else {
			content = "No content"
		}

		var location string
		if article.Location != nil {
			location = article.Location.Name
		} else {
			location = "No location"
		}

		var topicName string
		if int(article.Topic) < len(topicNames) {
			topicName = topicNames[article.Topic]
		} else {
			topicName = fmt.Sprintf("Topic_%d", article.Topic)
		}

		var hasImage bool
		var imageSize int
		var imageCaption string
		if article.Thumbnail != nil {
			hasImage = true
			imageSize = len(article.Thumbnail.Image)
			imageCaption = article.Thumbnail.Caption
		}

		debugArticles = append(debugArticles, DebugArticle{
			Title:        article.Title,
			Content:      content,
			Topic:        topicName,
			Location:     location,
			HasImage:     hasImage,
			ImageSize:    imageSize,
			ImageCaption: imageCaption,
		})
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("debug/articles_%s.json", timestamp)

	// Save to JSON file
	jsonData, err := json.MarshalIndent(debugArticles, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling articles: %v\n", err)
		return
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing debug file: %v\n", err)
		return
	}

	fmt.Printf("Debug: Saved %d articles to %s\n", len(debugArticles), filename)
}
