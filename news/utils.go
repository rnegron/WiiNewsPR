package news

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/image/draw"

	"image"
	"image/jpeg"
	_ "image/png"
)

func HttpGet(url string, userAgent ...string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if len(userAgent) > 0 && userAgent[0] != "" {
		req.Header.Set("User-Agent", userAgent[0])
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func IsDuplicateArticle(previousArticles []string, currentArticle string) bool {
	for _, previousArticle := range previousArticles {
		diff := difflib.NewMatcher([]string{currentArticle}, []string{previousArticle})
		if diff.QuickRatio() >= 0.85 {
			return true
		}
	}

	return false
}

func DownloadImage(imageURL string) ([]byte, error) {
	if imageURL == "" {
		return nil, fmt.Errorf("empty image URL")
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Limit image size
	const maxImageSize = 500 * 1024 // 500KB
	if len(data) > maxImageSize {
		return nil, fmt.Errorf("image too large: %d bytes", len(data))
	}

	return data, nil
}

func ConvertImage(data []byte) []byte {
	origImage, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	newImage := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), origImage, origImage.Bounds(), draw.Over, nil)

	var outputImgWriter bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&outputImgWriter), newImage, nil)
	if err != nil {
		return nil
	}

	return outputImgWriter.Bytes()
}

func CleanHTMLEntities(content string) string {
	content = html.UnescapeString(content)

	iframeRegex := regexp.MustCompile(`(?s)<iframe.*?>.*?</iframe>`)
	content = iframeRegex.ReplaceAllString(content, "")

	scriptRegex := regexp.MustCompile(`(?s)<script.*?>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	// Remove all HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	content = htmlTagRegex.ReplaceAllString(content, "")

	// RTVE specific tag that calls on another article
	noticiaRegex := regexp.MustCompile(`@@NOTICIA\[[^\]]*\]`)
	content = noticiaRegex.ReplaceAllString(content, "")

	fotoRegex := regexp.MustCompile(`@@FOTO\[[^\]]*\]`)
	content = fotoRegex.ReplaceAllString(content, "")

	mediaRegex := regexp.MustCompile(`@@MEDIA\[[^\]]*\]`)
	content = mediaRegex.ReplaceAllString(content, "")

	replacements := map[string]string{
		"&nbsp;":   " ",
		"&lt;":     "<",
		"&gt;":     ">",
		"&amp;":    "&",
		"&quot;":   "\"",
		"&apos;":   "'",
		"&uacute;": "ú",
		"&iacute;": "í",
		"&oacute;": "ó",
		"&aacute;": "á",
		"&eacute;": "é",
		"&ntilde;": "ñ",
		"&Uacute;": "Ú",
		"&Iacute;": "Í",
		"&Oacute;": "Ó",
		"&Aacute;": "Á",
		"&Eacute;": "É",
		"&Ntilde;": "Ñ",
		"&uuml;":   "ü",
		"&Uuml;":   "Ü",
	}

	for entity, char := range replacements {
		content = strings.ReplaceAll(content, entity, char)
	}

	return strings.TrimSpace(content)
}
