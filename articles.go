package main

import (
	"math"
	"unicode/utf16"
)

type Article struct {
	ID                uint32
	SourceIndex       uint32
	LocationIndex     uint32
	PictureTimestamp  uint32
	PictureIndex      uint32
	PublishedTime     uint32
	UpdatedTime       uint32
	HeadlineSize      uint32
	HeadlineOffset    uint32
	ArticleTextSize   uint32
	ArticleTextOffset uint32
}

type Image struct {
	CreditSize    uint32
	CreditOffset  uint32
	CaptionSize   uint32
	CaptionOffset uint32
	PictureSize   uint32
	PictureOffset uint32
}

func (n *News) MakeArticleTable() {
	n.Header.ArticleTableOffset = n.GetCurrentSize()

	// First write all metadata
	for i, article := range n.articles {
		publishedTime := currentTime

		// FORK UPDATE: Always use San Juan location (index 0)
		locationIndex := uint32(0)

		n.Articles = append(n.Articles, Article{
			ID:                uint32(i + 1),
			SourceIndex:       0,
			LocationIndex:     locationIndex,
			PictureTimestamp:  0,
			PictureIndex:      math.MaxUint32,
			PublishedTime:     fixTime(publishedTime),
			UpdatedTime:       fixTime(currentTime),
			HeadlineSize:      0,
			HeadlineOffset:    0,
			ArticleTextSize:   0,
			ArticleTextOffset: 0,
		})

		n.timestamps[article.Topic+1] = append(n.timestamps[article.Topic+1], Timestamp{
			Time:          fixTime(currentTime),
			ArticleNumber: uint32(i + 1),
		})
	}

	// Next write the text
	for i, article := range n.articles {
		encodedTitle := utf16.Encode([]rune(article.Title))
		encodedArticle := utf16.Encode([]rune(*article.Content))

		n.Articles[i].HeadlineSize = uint32(len(encodedTitle) * 2)
		n.Articles[i].ArticleTextSize = uint32(len(encodedArticle) * 2)

		n.Articles[i].HeadlineOffset = n.GetCurrentSize()
		n.ArticleText = append(n.ArticleText, encodedTitle...)

		// Null terminator
		n.ArticleText = append(n.ArticleText, 0)

		for n.GetCurrentSize()%4 != 0 {
			n.ArticleText = append(n.ArticleText, uint16(0))
		}

		n.Articles[i].ArticleTextOffset = n.GetCurrentSize()
		n.ArticleText = append(n.ArticleText, encodedArticle...)

		// Null terminator
		n.ArticleText = append(n.ArticleText, 0)

		for n.GetCurrentSize()%4 != 0 {
			n.ArticleText = append(n.ArticleText, uint16(0))
		}
	}

	n.Header.NumberOfArticles = uint32(len(n.Articles))
}

func (n *News) WriteImages() {
	n.Header.ImagesTableOffset = n.GetCurrentSize()

	// First, create a consistent list of articles with valid images
	var articlesWithImages []int // Store indices of articles that have valid images
	for i, article := range n.articles {
		if article.Thumbnail != nil && len(article.Thumbnail.Image) > 0 {
			articlesWithImages = append(articlesWithImages, i)
		}
	}

	// Create Image structs for articles with valid images
	for _, articleIndex := range articlesWithImages {
		article := n.articles[articleIndex]
		n.Images = append(n.Images, Image{
			CreditSize:    0,
			CreditOffset:  0,
			CaptionSize:   0,
			CaptionOffset: 0,
			PictureSize:   uint32(len(article.Thumbnail.Image)),
			PictureOffset: 0,
		})
	}

	for imageIndex, articleIndex := range articlesWithImages {
		article := n.articles[articleIndex]

		n.Images[imageIndex].PictureOffset = n.GetCurrentSize()
		n.ImagesData = append(n.ImagesData, article.Thumbnail.Image...)
		for n.GetCurrentSize()%4 != 0 {
			n.ImagesData = append(n.ImagesData, 0)
		}

		n.Articles[articleIndex].PictureIndex = uint32(imageIndex)
		n.Articles[articleIndex].PictureTimestamp = fixTime(currentTime)
	}

	for imageIndex, articleIndex := range articlesWithImages {
		article := n.articles[articleIndex]

		// Only process caption if it exists
		if article.Thumbnail.Caption != "" {
			caption := utf16.Encode([]rune(article.Thumbnail.Caption))
			n.Images[imageIndex].CaptionOffset = n.GetCurrentSize()
			n.Images[imageIndex].CaptionSize = uint32(len(caption) / 2)
			n.CaptionData = append(n.CaptionData, caption...)
			n.CaptionData = append(n.CaptionData, 0)

			for n.GetCurrentSize()%4 != 0 {
				n.CaptionData = append(n.CaptionData, uint16(0))
			}
		}
	}

	n.Header.NumberOfImages = uint32(len(n.Images))
}
