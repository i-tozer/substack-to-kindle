package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"substack-to-kindle/pkg/converter"
	"substack-to-kindle/pkg/pdfconverter"
	"substack-to-kindle/pkg/scraper"
	"substack-to-kindle/pkg/sender"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Parse command line arguments
	urlFlag := flag.String("url", "", "URL of the Substack article to convert")
	pdfFlag := flag.String("pdf", "", "Path to a local PDF file to convert")
	format := flag.String("format", "epub", "Output format: epub, azw3, or mobi")
	flag.Parse()

	// Validate format
	*format = strings.ToLower(*format)
	if *format != "epub" && *format != "azw3" && *format != "mobi" {
		log.Fatal("Format must be either 'epub', 'azw3', or 'mobi'")
	}

	// Warn if MOBI format is selected
	if *format == "mobi" {
		log.Println("Warning: MOBI format is no longer supported by Amazon's Send to Kindle service. Consider using EPUB or AZW3 instead.")
	}

	var result *converter.ConversionResult

	// Check if PDF file is provided
	if *pdfFlag != "" {
		// Process PDF file
		fmt.Println("Processing PDF file:", *pdfFlag)

		// Validate PDF path
		pdfPath, err := filepath.Abs(*pdfFlag)
		if err != nil {
			log.Fatalf("Invalid PDF path: %v", err)
		}

		// Convert PDF to the specified format
		fmt.Printf("Converting PDF to %s format...\n", strings.ToUpper(*format))

		switch *format {
		case "epub":
			result, err = pdfconverter.ConvertPDFToEPUB(pdfPath)
		case "azw3":
			result, err = pdfconverter.ConvertPDFToAZW3(pdfPath)
		case "mobi":
			result, err = pdfconverter.ConvertPDFToMOBI(pdfPath)
		}

		if err != nil {
			log.Fatalf("Failed to convert PDF: %v", err)
		}

		fmt.Printf("Conversion successful: %s\n", result.FilePath)
	} else {
		// Check if URL is provided
		articleURL := *urlFlag
		if articleURL == "" {
			// Check if URL is provided as a positional argument
			if len(flag.Args()) > 0 {
				articleURL = flag.Args()[0]
			} else {
				log.Fatal("Please provide either a Substack article URL using the -url flag or a PDF file using the -pdf flag")
			}
		}

		// Validate URL
		parsedURL, err := url.Parse(articleURL)
		if err != nil {
			log.Fatalf("Invalid URL: %v", err)
		}

		// Check if it's a Substack URL
		host := parsedURL.Host
		if !strings.HasSuffix(host, "substack.com") && !strings.Contains(host, ".substack.") {
			log.Fatal("The URL must be from a Substack site")
		}

		// Step 1: Scrape the article
		fmt.Println("Scraping article from:", articleURL)
		article, err := scraper.ScrapeSubstack(articleURL)
		if err != nil {
			log.Fatalf("Failed to scrape article: %v", err)
		}
		fmt.Printf("Successfully scraped article: %s by %s\n", article.Title, article.Author)

		// Step 2: Convert to the specified format
		fmt.Printf("Converting article to %s format...\n", strings.ToUpper(*format))

		switch *format {
		case "epub":
			result, err = converter.ConvertToEPUB(article)
		case "azw3":
			result, err = converter.ConvertToAZW3(article)
		case "mobi":
			result, err = converter.ConvertToMOBI(article)
		}

		if err != nil {
			log.Fatalf("Failed to convert article: %v", err)
		}

		fmt.Printf("Conversion successful: %s\n", result.FilePath)
	}

	// Step 3: Send to Kindle
	fmt.Println("Sending to Kindle...")
	config := sender.LoadEmailConfigFromEnv()
	err = sender.SendToKindle(result, config)
	if err != nil {
		log.Fatalf("Failed to send to Kindle: %v", err)
	}
	fmt.Println("Successfully sent to Kindle!")

	// Clean up temporary files
	os.Remove(result.FilePath)
	fmt.Println("Temporary files cleaned up.")
}
