package pdfconverter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"substack-to-kindle/pkg/converter"
)

// ConvertPDFToEPUB converts a local PDF file to EPUB format
func ConvertPDFToEPUB(pdfPath string) (*converter.ConversionResult, error) {
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

	// Try to use Calibre's ebook-convert for conversion (better quality)
	if isEbookConvertAvailable() {
		fmt.Println("Converting PDF to EPUB using Calibre...")
		if err := convertWithCalibre(pdfPath, epubPath); err != nil {
			return nil, fmt.Errorf("failed to convert PDF to EPUB using Calibre: %w", err)
		}
	} else {
		// Fallback to alternative conversion method
		fmt.Println("Calibre not available. Using alternative conversion method...")
		if err := convertWithAlternative(pdfPath, epubPath); err != nil {
			return nil, fmt.Errorf("failed to convert PDF to EPUB: %w", err)
		}
	}

	// Create the conversion result
	result := &converter.ConversionResult{
		FilePath: epubPath,
		Title:    fileNameWithoutExt,
		Author:   "PDF Conversion", // Default author for PDF conversions
	}

	return result, nil
}

// ConvertPDFToAZW3 converts a local PDF file to AZW3 format
func ConvertPDFToAZW3(pdfPath string) (*converter.ConversionResult, error) {
	// First convert to EPUB
	epubResult, err := ConvertPDFToEPUB(pdfPath)
	if err != nil {
		return nil, err
	}

	// Then convert EPUB to AZW3
	tempDir := filepath.Dir(epubResult.FilePath)
	fileNameWithoutExt := strings.TrimSuffix(filepath.Base(epubResult.FilePath), ".epub")
	azw3Path := filepath.Join(tempDir, fileNameWithoutExt+".azw3")

	// Try to use Calibre for conversion
	if isEbookConvertAvailable() {
		fmt.Println("Converting EPUB to AZW3 using Calibre...")
		err := convertWithCalibre(epubResult.FilePath, azw3Path)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EPUB to AZW3: %w", err)
		}
	} else {
		// If Calibre is not available, use our own conversion method
		fmt.Println("Calibre not available. Using alternative conversion method...")
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
func ConvertPDFToMOBI(pdfPath string) (*converter.ConversionResult, error) {
	// First convert to EPUB
	epubResult, err := ConvertPDFToEPUB(pdfPath)
	if err != nil {
		return nil, err
	}

	// Then convert EPUB to MOBI
	tempDir := filepath.Dir(epubResult.FilePath)
	fileNameWithoutExt := strings.TrimSuffix(filepath.Base(epubResult.FilePath), ".epub")
	mobiPath := filepath.Join(tempDir, fileNameWithoutExt+".mobi")

	// Try to use Calibre for conversion
	if isEbookConvertAvailable() {
		fmt.Println("Converting EPUB to MOBI using Calibre...")
		err := convertWithCalibre(epubResult.FilePath, mobiPath)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EPUB to MOBI: %w", err)
		}
	} else {
		// If Calibre is not available, use our own conversion method
		fmt.Println("Calibre not available. Using alternative conversion method...")
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

// convertWithAlternative uses an alternative method to convert PDF to EPUB
func convertWithAlternative(pdfPath, epubPath string) error {
	// This is a placeholder for an alternative conversion method
	// In a real implementation, you would use a library like pdf2htmlEX,
	// poppler-utils, or a Go PDF library to extract text and images,
	// then build an EPUB file using the go-epub library.

	// For now, we'll return an error since we don't have a built-in alternative
	return fmt.Errorf("no alternative PDF to EPUB conversion method available; please install Calibre")
}

// convertToAZW3 converts an EPUB file to AZW3 format
func convertToAZW3(epubPath, azw3Path string) error {
	// This is a placeholder for direct EPUB to AZW3 conversion
	// In a real implementation, you would use a library to convert EPUB to AZW3

	// For now, we'll return an error since we don't have a built-in alternative
	return fmt.Errorf("no alternative EPUB to AZW3 conversion method available; please install Calibre")
}

// convertToMOBI converts an EPUB file to MOBI format
func convertToMOBI(epubPath, mobiPath string) error {
	// This is a placeholder for direct EPUB to MOBI conversion
	// In a real implementation, you would use a library to convert EPUB to MOBI

	// For now, we'll return an error since we don't have a built-in alternative
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
