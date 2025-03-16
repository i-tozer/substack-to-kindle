# Substack to Kindle

A Go application that converts Substack articles to EPUB, AZW3, or MOBI format and sends them to your Kindle device via email. It can also convert local PDF files to Kindle-compatible formats.

## Prerequisites

- Go 1.21 or higher
- Gmail account (or other email provider with SMTP access)

### Calibre (Optional)

Calibre is optional and only used if explicitly requested:

- **macOS**: `brew install --cask calibre`
- **Linux**: `sudo apt-get install calibre`
- **Windows**: Download and install from [Calibre's website](https://calibre-ebook.com/download)

## Setup

1. Clone this repository
2. Install dependencies:
   ```
   go mod download
   ```
3. Copy `.env.example` to `.env` and configure it:
   - Set `EMAIL_FROM` to your email address (e.g., your.email@example.com)
   - Set `EMAIL_TO` to your Kindle email address (find this in your Amazon account under "Manage Your Content and Devices" > "Devices")
   - Set `EMAIL_PASSWORD` to your app password (see note below)
   - Set `SMTP_HOST` and `SMTP_PORT` according to your email provider

### Note on Gmail App Passwords

If you're using Gmail, you'll need to use an "App Password" instead of your regular password:

1. Enable 2-Step Verification on your Google account
2. Go to [App Passwords](https://myaccount.google.com/apppasswords)
3. Select "Mail" and your device
4. Generate and copy the 16-character password
5. Paste this password in your `.env` file as `EMAIL_PASSWORD`

### Kindle Email Setup

To receive documents on your Kindle:

1. Add your sending email address to your approved list in Amazon (under "Manage Your Content and Devices" > "Preferences" > "Personal Document Settings")
2. Use your Kindle email address as the `EMAIL_TO` value in your `.env` file

## Usage

### Converting Substack Articles

Convert and send a Substack article to your Kindle:

```
go run main.go -url https://example.substack.com/p/article-name
```

Or simply:

```
go run main.go https://example.substack.com/p/article-name
```

### Converting PDF Files

Convert and send a local PDF file to your Kindle:

```
go run main.go -pdf /path/to/your/file.pdf
```

#### PDF Conversion Options

The following options are available for PDF conversion:

- `-title "Your Custom Title"` - Set a custom title for the document
- `-author "Author Name"` - Set a custom author for the document
- `-skip-calibre=false` - Use Calibre if it's available (default is to skip Calibre)
- `-include-pdf=true` - Include the original PDF in the output file (default is false)

Example with options:

```
go run main.go -pdf /path/to/your/file.pdf -title "My Book" -author "John Doe" -format azw3 -include-pdf=true
```

### Specifying Output Format

By default, the application converts content to EPUB format, which is the recommended format for sending to Kindle devices. You can specify a different output format using the `-format` flag:

```
go run main.go -url https://example.substack.com/p/article-name -format azw3
```

Or for PDF files:

```
go run main.go -pdf /path/to/your/file.pdf -format azw3
```

> **Important Note**: As of 2023, Amazon no longer supports sending MOBI files through the Send to Kindle service. While the application still supports creating MOBI files, we recommend using EPUB or AZW3 formats for Kindle delivery.

#### Format Comparison

- **EPUB** (default): Universal format, works on most e-readers including Kindle (via email)
- **AZW3**: Amazon's proprietary format with better formatting and features
- **MOBI**: Amazon's older format, no longer supported by Send to Kindle service

### Example

```
go run main.go https://dominiccummings.substack.com/p/people-ideas-machines-i-notes-on
```

This will:
1. Scrape the article "'People, ideas, machines' I: Notes on 'Winning the Next War'" by Dominic Cummings
2. Convert it to EPUB format with proper formatting and images
3. Send it to your Kindle email address

## Features

- Scrapes Substack articles preserving formatting and images
- Converts local PDF files to Kindle-compatible formats
- Extracts text from PDFs for better reading experience
- Converts content to EPUB (default), AZW3, or MOBI format
- Direct conversion to AZW3 and MOBI formats without requiring Calibre
- Uses Calibre for conversion when available (better quality)
- Sends the converted file directly to your Kindle device
- Cleans up temporary files after sending

## How It Works

1. **Input Processing**:
   - For Substack URLs: Extracts article content, title, author, and images from the URL
   - For PDF files: Processes the local PDF file and extracts text content
2. **Conversion**: 
   - For EPUB: Converts the content directly to EPUB format
   - For AZW3/MOBI: Converts directly to the requested format
   - By default, uses built-in text extraction for PDFs
   - If requested with `-skip-calibre=false`, will try to use Calibre for conversion
3. **Delivery**: Sends the converted file to your Kindle email address

## Limitations

- Works only with public Substack articles (no paywall content)
- PDF conversion requires Calibre to be installed for best results
- Text extraction from PDFs may not preserve complex formatting or images
- Some complex formatting or interactive elements may not be preserved
- MOBI format is no longer supported by Amazon's Send to Kindle service

## Troubleshooting

### PDF Conversion Issues

If you encounter issues with the default PDF conversion method, try the following:

1. Include the original PDF in the output file:
   ```
   go run main.go -pdf /path/to/your/file.pdf -include-pdf=true
   ```

2. Try using Calibre for conversion (if installed):
   ```
   go run main.go -pdf /path/to/your/file.pdf -skip-calibre=false
   ```

3. Make sure your PDF is not password-protected or encrypted

4. For complex PDFs, consider pre-processing them with other tools before conversion

## Project Structure

- `main.go`: Main application entry point
- `pkg/scraper`: Module for extracting content from Substack articles
- `pkg/converter`: Module for converting articles to EPUB, AZW3, or MOBI format
- `pkg/pdfconverter`: Module for converting PDF files to Kindle-compatible formats
- `pkg/sender`: Module for sending files to Kindle via email 