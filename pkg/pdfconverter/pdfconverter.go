package pdfconverter

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"substack-to-kindle/pkg/converter"

	"github.com/bmaupin/go-epub"
	"github.com/ledongthuc/pdf"
)

// ConversionOptions contains options for PDF conversion
type ConversionOptions struct {
	// SkipCalibre skips using Calibre even if it's available
	SkipCalibre bool
	// CustomTitle overrides the filename as the title
	CustomTitle string
	// CustomAuthor overrides the default author
	CustomAuthor string
	// IncludeOriginalPDF includes the original PDF in the EPUB
	IncludeOriginalPDF bool
}

// DefaultOptions returns the default conversion options
func DefaultOptions() *ConversionOptions {
	return &ConversionOptions{
		SkipCalibre:        true,
		CustomTitle:        "",
		CustomAuthor:       "PDF Conversion",
		IncludeOriginalPDF: false,
	}
}

// ConvertPDFToEPUB converts a local PDF file to EPUB format
func ConvertPDFToEPUB(pdfPath string, options *ConversionOptions) (*converter.ConversionResult, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultOptions()
	}

	// Validate that the file exists and is a PDF
	if err := validatePDFFile(pdfPath); err != nil {
		return nil, err
	}

	// Create a temporary directory for our files
	tempDir, err := os.MkdirTemp("", "pdf-kindle-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Get the filename without extension
	baseName := filepath.Base(pdfPath)
	fileNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Sanitize the filename
	fileNameWithoutExt = sanitizeFilename(fileNameWithoutExt)

	// Create the output path
	epubPath := filepath.Join(tempDir, fileNameWithoutExt+".epub")

	// Determine title and author
	title := fileNameWithoutExt
	if options.CustomTitle != "" {
		title = options.CustomTitle
	}

	author := options.CustomAuthor

	// Try to use Calibre's ebook-convert for conversion (better quality)
	calibreSuccess := false
	if isEbookConvertAvailable() && !options.SkipCalibre {
		fmt.Println("Converting PDF to EPUB using Calibre...")
		err := convertWithCalibre(pdfPath, epubPath)
		if err != nil {
			fmt.Printf("Calibre conversion failed: %v\n", err)
			fmt.Println("Trying alternative conversion method...")
		} else {
			calibreSuccess = true
		}
	}

	// If Calibre failed or was skipped, use alternative method
	if !calibreSuccess {
		fmt.Println("Using alternative conversion method...")
		if err := convertWithAlternative(pdfPath, epubPath, title, author, options.IncludeOriginalPDF); err != nil {
			return nil, fmt.Errorf("failed to convert PDF to EPUB: %w", err)
		}
	}

	// Create the conversion result
	result := &converter.ConversionResult{
		FilePath: epubPath,
		Title:    title,
		Author:   author,
	}

	return result, nil
}

// ConvertPDFToAZW3 converts a local PDF file to AZW3 format
func ConvertPDFToAZW3(pdfPath string, options *ConversionOptions) (*converter.ConversionResult, error) {
	// First convert to EPUB
	epubResult, err := ConvertPDFToEPUB(pdfPath, options)
	if err != nil {
		return nil, err
	}

	// Then convert EPUB to AZW3
	tempDir := filepath.Dir(epubResult.FilePath)
	fileNameWithoutExt := strings.TrimSuffix(filepath.Base(epubResult.FilePath), ".epub")
	azw3Path := filepath.Join(tempDir, fileNameWithoutExt+".azw3")

	// Try to use Calibre for conversion
	calibreSuccess := false
	if isEbookConvertAvailable() && !options.SkipCalibre {
		fmt.Println("Converting EPUB to AZW3 using Calibre...")
		err := convertWithCalibre(epubResult.FilePath, azw3Path)
		if err != nil {
			fmt.Printf("Calibre conversion failed: %v\n", err)
			fmt.Println("Trying alternative conversion method...")
		} else {
			calibreSuccess = true
		}
	}

	// If Calibre failed or was skipped, use alternative method
	if !calibreSuccess {
		fmt.Println("Using alternative conversion method...")
		err := convertToAZW3(epubResult.FilePath, azw3Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EPUB to AZW3: %w", err)
		}
	}

	// Clean up the intermediate EPUB file
	os.Remove(epubResult.FilePath)

	// Create the conversion result
	result := &converter.ConversionResult{
		FilePath: azw3Path,
		Title:    epubResult.Title,
		Author:   epubResult.Author,
	}

	return result, nil
}

// ConvertPDFToMOBI converts a local PDF file to MOBI format
func ConvertPDFToMOBI(pdfPath string, options *ConversionOptions) (*converter.ConversionResult, error) {
	// First convert to EPUB
	epubResult, err := ConvertPDFToEPUB(pdfPath, options)
	if err != nil {
		return nil, err
	}

	// Then convert EPUB to MOBI
	tempDir := filepath.Dir(epubResult.FilePath)
	fileNameWithoutExt := strings.TrimSuffix(filepath.Base(epubResult.FilePath), ".epub")
	mobiPath := filepath.Join(tempDir, fileNameWithoutExt+".mobi")

	// Try to use Calibre for conversion
	calibreSuccess := false
	if isEbookConvertAvailable() && !options.SkipCalibre {
		fmt.Println("Converting EPUB to MOBI using Calibre...")
		err := convertWithCalibre(epubResult.FilePath, mobiPath)
		if err != nil {
			fmt.Printf("Calibre conversion failed: %v\n", err)
			fmt.Println("Trying alternative conversion method...")
		} else {
			calibreSuccess = true
		}
	}

	// If Calibre failed or was skipped, use alternative method
	if !calibreSuccess {
		fmt.Println("Using alternative conversion method...")
		err := convertToMOBI(epubResult.FilePath, mobiPath)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EPUB to MOBI: %w", err)
		}
	}

	// Clean up the intermediate EPUB file
	os.Remove(epubResult.FilePath)

	// Create the conversion result
	result := &converter.ConversionResult{
		FilePath: mobiPath,
		Title:    epubResult.Title,
		Author:   epubResult.Author,
	}

	return result, nil
}

// validatePDFFile checks if the file exists and is a PDF
func validatePDFFile(pdfPath string) error {
	// Check if file exists
	info, err := os.Stat(pdfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", pdfPath)
		}
		return fmt.Errorf("error accessing file: %w", err)
	}

	// Check if it's a regular file
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", pdfPath)
	}

	// Check if it has a .pdf extension
	if !strings.HasSuffix(strings.ToLower(pdfPath), ".pdf") {
		return fmt.Errorf("file does not have a .pdf extension: %s", pdfPath)
	}

	return nil
}

// isEbookConvertAvailable checks if Calibre's ebook-convert is available
func isEbookConvertAvailable() bool {
	_, err := exec.LookPath("ebook-convert")
	return err == nil
}

// convertWithCalibre converts a file using Calibre's ebook-convert
func convertWithCalibre(inputPath, outputPath string) error {
	cmd := exec.Command("ebook-convert", inputPath, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebook-convert failed: %w, output: %s", err, output)
	}
	return nil
}

// extractTextFromPDF extracts text from a PDF file
func extractTextFromPDF(pdfPath string) (string, error) {
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", fmt.Errorf("failed to read text from PDF: %w", err)
	}

	return buf.String(), nil
}

// cleanText cleans and sanitizes text for HTML
func cleanText(text string) string {
	// Escape HTML special characters
	text = html.EscapeString(text)

	// Replace multiple spaces with a single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Remove any control characters
	re = regexp.MustCompile(`[\x00-\x1F\x7F]`)
	text = re.ReplaceAllString(text, "")

	return text
}

// convertWithAlternative uses an alternative method to convert PDF to EPUB
func convertWithAlternative(pdfPath, epubPath, title, author string, includeOriginalPDF bool) error {
	// Create a basic EPUB with a note about the PDF
	e := epub.NewEpub(title)
	e.SetAuthor(author)

	// Extract text from PDF
	fmt.Println("Extracting text from PDF...")
	pdfText, err := extractTextFromPDF(pdfPath)
	if err != nil {
		fmt.Printf("Warning: Failed to extract text from PDF: %v\n", err)
		pdfText = "Failed to extract text from this PDF. The original PDF file has been included as an attachment."
	}

	// Add a cover page
	coverContent := fmt.Sprintf(`
		<html>
			<head>
				<title>%s</title>
			</head>
			<body>
				<h1>%s</h1>
				<h2>By %s</h2>
				<p>This is a converted PDF document.</p>
				<p>The original PDF may contain formatting and content that could not be fully preserved in this conversion.</p>
			</body>
		</html>
	`, html.EscapeString(title), html.EscapeString(title), html.EscapeString(author))

	_, err = e.AddSection(coverContent, "Cover", "", "")
	if err != nil {
		return fmt.Errorf("failed to add cover page: %w", err)
	}

	// Format the extracted text into HTML
	// Split the text into paragraphs
	paragraphs := strings.Split(pdfText, "\n\n")

	// Create HTML content with paragraphs
	var contentBuilder strings.Builder
	contentBuilder.WriteString("<html><head><title>PDF Content</title></head><body>")

	// Process paragraphs in chunks to avoid creating too large HTML sections
	const maxParagraphsPerSection = 100
	numSections := (len(paragraphs) + maxParagraphsPerSection - 1) / maxParagraphsPerSection

	for sectionIdx := 0; sectionIdx < numSections; sectionIdx++ {
		var sectionBuilder strings.Builder
		sectionBuilder.WriteString("<html><head><title>PDF Content</title></head><body>")

		start := sectionIdx * maxParagraphsPerSection
		end := (sectionIdx + 1) * maxParagraphsPerSection
		if end > len(paragraphs) {
			end = len(paragraphs)
		}

		for _, paragraph := range paragraphs[start:end] {
			// Skip empty paragraphs
			paragraph = strings.TrimSpace(paragraph)
			if paragraph == "" {
				continue
			}

			// Replace single newlines with spaces
			paragraph = strings.ReplaceAll(paragraph, "\n", " ")

			// Clean and sanitize the text
			paragraph = cleanText(paragraph)

			// Add the paragraph to the HTML
			sectionBuilder.WriteString("<p>" + paragraph + "</p>")
		}

		sectionBuilder.WriteString("</body></html>")

		// Add the section to the EPUB
		sectionTitle := fmt.Sprintf("Content Part %d", sectionIdx+1)
		_, err = e.AddSection(sectionBuilder.String(), sectionTitle, "", "")
		if err != nil {
			return fmt.Errorf("failed to add content section %d: %w", sectionIdx+1, err)
		}
	}

	// Include the original PDF if requested
	if includeOriginalPDF {
		pdfFileName := filepath.Base(pdfPath)

		// Add the PDF as an image (it will be stored as a binary file in the EPUB)
		pdfImagePath, err := e.AddImage(pdfPath, pdfFileName)
		if err != nil {
			fmt.Printf("Warning: Failed to include original PDF: %v\n", err)
		} else {
			// Add a section with information about the PDF
			pdfSection := fmt.Sprintf(`
				<html>
					<head>
						<title>Original PDF</title>
					</head>
					<body>
						<h1>Original PDF Document</h1>
						<p>The original PDF file "%s" has been included as an attachment.</p>
						<p>Some e-readers may allow you to open this PDF directly.</p>
						<p>If your e-reader supports it, you can <a href="%s">click here to open the PDF</a>.</p>
					</body>
				</html>
			`, html.EscapeString(pdfFileName), html.EscapeString(pdfImagePath))

			_, err = e.AddSection(pdfSection, "Original PDF", "", "")
			if err != nil {
				fmt.Printf("Warning: Failed to add PDF section: %v\n", err)
			}
		}
	}

	// Write the EPUB file
	err = e.Write(epubPath)
	if err != nil {
		return fmt.Errorf("failed to write EPUB file: %w", err)
	}

	return nil
}

// convertToAZW3 converts an EPUB file to AZW3 format
func convertToAZW3(epubPath, azw3Path string) error {
	// For now, we'll return an error since we don't have a built-in alternative
	// In a real implementation, you would use a library to convert EPUB to AZW3
	return fmt.Errorf("no alternative EPUB to AZW3 conversion method available; please install Calibre")
}

// convertToMOBI converts an EPUB file to MOBI format
func convertToMOBI(epubPath, mobiPath string) error {
	// For now, we'll return an error since we don't have a built-in alternative
	// In a real implementation, you would use a library to convert EPUB to MOBI
	return fmt.Errorf("no alternative EPUB to MOBI conversion method available; please install Calibre")
}

// sanitizeFilename sanitizes a filename to be safe for use in a file path
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
