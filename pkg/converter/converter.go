package converter

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"substack-to-kindle/pkg/scraper"

	"github.com/bmaupin/go-epub"
	"github.com/leotaku/mobi"
	"golang.org/x/text/language"
)

// OutputFormat represents the output format for the conversion
type OutputFormat string

const (
	// FormatEPUB represents the EPUB format
	FormatEPUB OutputFormat = "epub"
	// FormatAZW3 represents the AZW3 format (Kindle)
	FormatAZW3 OutputFormat = "azw3"
	// FormatMOBI represents the MOBI format (Kindle)
	FormatMOBI OutputFormat = "mobi"
)

// ConversionResult contains information about the converted file
type ConversionResult struct {
	FilePath string
	Title    string
	Author   string
}

// ConvertArticle converts a Substack article to the specified format
func ConvertArticle(article *scraper.Article, format OutputFormat) (*ConversionResult, error) {
	// Create a temporary directory for our files
	tempDir, err := os.MkdirTemp("", "substack-kindle-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s - %s",
		sanitizeFilename(article.Title),
		sanitizeFilename(article.Author))

	var outputPath string

	// For EPUB format
	if format == FormatEPUB {
		fmt.Println("Creating EPUB file...")
		epubPath, err := createEPUB(article, tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create EPUB: %w", err)
		}
		outputPath = epubPath
	} else if format == FormatAZW3 {
		// Try using Calibre first (better quality conversion)
		if isEbookConvertAvailable() {
			fmt.Println("Creating EPUB file...")
			epubPath, err := createEPUB(article, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to create EPUB: %w", err)
			}

			fmt.Println("Converting from EPUB to AZW3 format using Calibre...")
			outputPath, err = convertToFormat(epubPath, "azw3")
			if err != nil {
				log.Printf("Warning: Failed to convert to AZW3 using Calibre: %v. Trying direct conversion...", err)
				// Fall back to direct conversion
				outputPath = ""
			} else {
				// Clean up the intermediate EPUB file
				os.Remove(epubPath)
			}
		}

		// If Calibre conversion failed or not available, use direct conversion
		if outputPath == "" {
			fmt.Println("Creating AZW3 file directly...")
			azw3Path := filepath.Join(tempDir, filename+".azw3")
			err := createAZW3(article, azw3Path)
			if err != nil {
				return nil, fmt.Errorf("failed to create AZW3: %w", err)
			}
			outputPath = azw3Path
		}
	} else if format == FormatMOBI {
		// Try using Calibre first (better quality conversion)
		if isEbookConvertAvailable() {
			fmt.Println("Creating EPUB file...")
			epubPath, err := createEPUB(article, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to create EPUB: %w", err)
			}

			fmt.Println("Converting from EPUB to MOBI format using Calibre...")
			outputPath, err = convertToFormat(epubPath, "mobi")
			if err != nil {
				log.Printf("Warning: Failed to convert to MOBI using Calibre: %v. Trying direct conversion...", err)
				// Fall back to direct conversion
				outputPath = ""
			} else {
				// Clean up the intermediate EPUB file
				os.Remove(epubPath)
			}
		}

		// If Calibre conversion failed or not available, use direct conversion
		if outputPath == "" {
			fmt.Println("Creating MOBI file directly...")
			mobiPath := filepath.Join(tempDir, filename+".mobi")
			err := createMOBI(article, mobiPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create MOBI: %w", err)
			}
			outputPath = mobiPath
		}
	} else {
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}

	return &ConversionResult{
		FilePath: outputPath,
		Title:    article.Title,
		Author:   article.Author,
	}, nil
}

// ConvertToEPUB converts a Substack article to EPUB format
// Note: Kindle can accept EPUB files directly via email
func ConvertToEPUB(article *scraper.Article) (*ConversionResult, error) {
	return ConvertArticle(article, FormatEPUB)
}

// ConvertToAZW3 converts a Substack article to AZW3 format
func ConvertToAZW3(article *scraper.Article) (*ConversionResult, error) {
	return ConvertArticle(article, FormatAZW3)
}

// ConvertToMOBI converts a Substack article to MOBI format
func ConvertToMOBI(article *scraper.Article) (*ConversionResult, error) {
	return ConvertArticle(article, FormatMOBI)
}

// isEbookConvertAvailable checks if Calibre's ebook-convert tool is available
func isEbookConvertAvailable() bool {
	_, err := exec.LookPath("ebook-convert")
	return err == nil
}

// convertToFormat converts an EPUB file to the specified format using Calibre
func convertToFormat(epubPath, format string) (string, error) {
	// Generate output path
	outputPath := strings.TrimSuffix(epubPath, ".epub") + "." + format

	// Run ebook-convert command
	cmd := exec.Command("ebook-convert", epubPath, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to convert to %s: %w, output: %s", format, err, output)
	}

	return outputPath, nil
}

// createAZW3 creates an AZW3 file directly from the article using the leotaku/mobi library
func createAZW3(article *scraper.Article, outputPath string) error {
	return createMobiFormat(article, outputPath, "azw3")
}

// createMOBI creates a MOBI file directly from the article using the leotaku/mobi library
func createMOBI(article *scraper.Article, outputPath string) error {
	return createMobiFormat(article, outputPath, "mobi")
}

// createMobiFormat creates a MOBI or AZW3 file directly from the article
func createMobiFormat(article *scraper.Article, outputPath, format string) error {
	// Download images to temporary directory
	tempDir := filepath.Dir(outputPath)
	imageMap := make(map[string]string)

	for _, imgURL := range article.ImageURLs {
		imgPath, err := downloadImage(imgURL, tempDir)
		if err != nil {
			continue // Skip this image if download fails
		}
		imageMap[imgURL] = imgPath
	}

	// Replace image URLs in content with local file references
	content := article.Content
	for origURL, localPath := range imageMap {
		content = strings.ReplaceAll(content, origURL, filepath.Base(localPath))
	}

	// Create HTML content
	htmlContent := fmt.Sprintf(`
		<html>
		<head>
			<title>%s</title>
			<style>
				body {
					font-family: serif;
					margin: 5%%;
					text-align: justify;
				}
				h1, h2, h3, h4, h5, h6 {
					text-align: left;
					margin-top: 1em;
				}
				img {
					max-width: 100%%;
					height: auto;
				}
				blockquote {
					margin: 1em 2em;
					font-style: italic;
				}
			</style>
		</head>
		<body>
			<h1>%s</h1>
			<p><strong>By %s</strong></p>
			<p><em>Published: %s</em></p>
			<p><em>Source: <a href="%s">%s</a></em></p>
			<hr/>
			%s
		</body>
		</html>
	`,
		article.Title,
		article.Title,
		article.Author,
		article.PublishedAt.Format("January 2, 2006"),
		article.URL,
		article.URL,
		content,
	)

	// Create a chapter with the article content
	ch := mobi.Chapter{
		Title:  article.Title,
		Chunks: mobi.Chunks(htmlContent),
	}

	// Create the book
	mb := mobi.Book{
		Title:       article.Title,
		Authors:     []string{article.Author},
		CreatedDate: time.Now(),
		Language:    language.English,
		Chapters:    []mobi.Chapter{ch},
		UniqueID:    rand.Uint32(),
	}

	// Convert book to PalmDB database
	db := mb.Realize()

	// Write database to file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	err = db.Write(f)
	if err != nil {
		return fmt.Errorf("failed to write %s file: %w", format, err)
	}

	return nil
}

// createEPUB creates an EPUB file from the article
func createEPUB(article *scraper.Article, tempDir string) (string, error) {
	// Create a new EPUB
	e := epub.NewEpub(article.Title)
	e.SetAuthor(article.Author)

	// Download and add images
	imageMap := make(map[string]string)
	for _, imgURL := range article.ImageURLs {
		imgPath, err := downloadImage(imgURL, tempDir)
		if err != nil {
			continue // Skip this image if download fails
		}

		// Add image to EPUB
		imgFilename := filepath.Base(imgPath)
		internalPath, err := e.AddImage(imgPath, imgFilename)
		if err != nil {
			continue
		}

		// Map original URL to internal EPUB path
		imageMap[imgURL] = internalPath
	}

	// Replace image URLs in content
	content := article.Content
	for origURL, epubPath := range imageMap {
		content = strings.ReplaceAll(content, origURL, epubPath)
	}

	// Add CSS
	cssContent := `
		body {
			font-family: serif;
			margin: 5%;
			text-align: justify;
		}
		h1, h2, h3, h4, h5, h6 {
			text-align: left;
			margin-top: 1em;
		}
		img {
			max-width: 100%;
			height: auto;
		}
		blockquote {
			margin: 1em 2em;
			font-style: italic;
		}
	`

	// Create a temporary CSS file
	cssFile, err := os.CreateTemp(tempDir, "style-*.css")
	if err != nil {
		return "", fmt.Errorf("failed to create CSS file: %w", err)
	}
	defer cssFile.Close()

	_, err = cssFile.WriteString(cssContent)
	if err != nil {
		return "", fmt.Errorf("failed to write CSS content: %w", err)
	}

	// Add the CSS file to the EPUB
	cssPath, err := e.AddCSS(cssFile.Name(), "style.css")
	if err != nil {
		return "", fmt.Errorf("failed to add CSS: %w", err)
	}

	// Create HTML content with metadata
	htmlContent := fmt.Sprintf(`
		<html>
		<head>
			<title>%s</title>
			<link rel="stylesheet" type="text/css" href="%s" />
		</head>
		<body>
			<h1>%s</h1>
			<p><strong>By %s</strong></p>
			<p><em>Published: %s</em></p>
			<p><em>Source: <a href="%s">%s</a></em></p>
			<hr/>
			%s
		</body>
		</html>
	`,
		article.Title,
		cssPath,
		article.Title,
		article.Author,
		article.PublishedAt.Format("January 2, 2006"),
		article.URL,
		article.URL,
		content,
	)

	// Add the section with content
	_, err = e.AddSection(htmlContent, article.Title, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to add content: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s - %s.epub",
		sanitizeFilename(article.Title),
		sanitizeFilename(article.Author))
	epubPath := filepath.Join(tempDir, filename)

	// Write EPUB to file
	err = e.Write(epubPath)
	if err != nil {
		return "", fmt.Errorf("failed to write EPUB: %w", err)
	}

	return epubPath, nil
}

// downloadImage downloads an image from a URL to the temp directory
func downloadImage(url string, tempDir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Create a file to save the image
	filename := filepath.Base(url)
	if filename == "" || !strings.Contains(filename, ".") {
		filename = fmt.Sprintf("image_%d.jpg", time.Now().UnixNano())
	}

	imgPath := filepath.Join(tempDir, filename)
	file, err := os.Create(imgPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy the image data to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return imgPath, nil
}

// sanitizeFilename removes invalid characters from a filename
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim spaces and limit length
	result = strings.TrimSpace(result)
	if len(result) > 100 {
		result = result[:100]
	}

	return result
}
