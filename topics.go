package main

import (
	"WiiNewsPR/news"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf16"
)

// FORK UPDATE: Since we only support English, we can hardcode the topics and their text here
var topics = []string{"National News", "International News", "Sports", "Entertainment", "Business", "Science", "Technology"}

type Topic struct {
	TextOffset           uint32
	NumberOfArticles     uint32
	TimestampTableOffset uint32
}

// Timestamp handles the time an article was obtained.
type Timestamp struct {
	Time          uint32
	ArticleNumber uint32
}

// NewsCache contains the bare minimum for articles we grabbed in the past.
type NewsCache struct {
	ID        uint32     `json:"id"`
	Timestamp uint32     `json:"timestamp"`
	Topic     news.Topic `json:"topic"`
	Title     string     `json:"title"`
}

// ReadNewsCache creates the topic table as well as the timestamp table for articles.
// This is quite an annoying job as for some reason it needs to make the timestamp table for every single article, even ones
// from past hours. Due to this we are required to cache what articles we used.
func (n *News) ReadNewsCache(cacheDir string) {
	topicsLength := len(topics) + 1

	n.topics = make([]Topic, topicsLength)
	n.timestamps = make([][]Timestamp, topicsLength)

	for i := 0; i < 24; i++ {
		// Don't process the cache for the current hour.
		if i == n.currentHour {
			continue
		}

		var _articles []NewsCache
		inputFile := filepath.Join(cacheDir, fmt.Sprintf("cache_%d.news", i))
		data, err := os.ReadFile(inputFile)
		if err != nil {
			continue
		}

		err = json.Unmarshal(data, &_articles)
		checkError(err)

		for _, article := range _articles {
			n.topics[article.Topic+1].NumberOfArticles++
			n.oldArticleTitles = append(n.oldArticleTitles, article.Title)
			n.timestamps[article.Topic+1] = append(n.timestamps[article.Topic+1], Timestamp{
				Time:          article.Timestamp,
				ArticleNumber: article.ID,
			})
		}
	}
}

func (n *News) MakeTopicTable() {
	// Move the placeholder into the field being written.
	n.Header.TopicTableOffset = n.GetCurrentSize()
	n.Topics = n.topics

	topicsLength := len(topics) + 1
	n.Header.NumberOfTopics = uint32(topicsLength)

	// Now we copy all our data into the struct
	for i := 1; i < topicsLength; i++ {
		n.Topics[i].TimestampTableOffset = n.GetCurrentSize()
		n.Topics[i].NumberOfArticles = uint32(len(n.timestamps[i]))
		n.Timestamps = append(n.Timestamps, n.timestamps[i]...)
	}

	for i, topic := range topics {
		n.Topics[i+1].TextOffset = n.GetCurrentSize()
		n.TopicText = append(n.TopicText, utf16.Encode([]rune(topic))...)
		n.TopicText = append(n.TopicText, uint16(0))
	}
}

// WriteNewsCache writes the found articles for the current hour.
func (n *News) WriteNewsCache(cacheDir string) {
	// FORK UPDATE: create cache directory if it doesn't exist
	err := os.MkdirAll(cacheDir, os.ModePerm)
	checkError(err)

	// Order everything into the NewsCache struct
	var cache []NewsCache
	for i, article := range n.articles {
		cache = append(cache, NewsCache{
			ID:        n.Articles[i].ID,
			Timestamp: fixTime(currentTime),
			Topic:     article.Topic,
			Title:     article.Title,
		})
	}

	// Encode NewsCache array
	data, err := json.Marshal(cache)
	checkError(err)

	outputFile := filepath.Join(cacheDir, fmt.Sprintf("cache_%d.news", n.currentHour))
	err = os.WriteFile(outputFile, data, 0666)
	checkError(err)
}
