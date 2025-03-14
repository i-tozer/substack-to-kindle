# Substack to Kindle

A Go application that converts Substack articles to EPUB, AZW3, or MOBI format and sends them to your Kindle device via email.

## Prerequisites

- Go 1.21 or higher
- Gmail account (or other email provider with SMTP access)

### Calibre (Optional)

Calibre is optional but recommended for better quality conversions:

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

Convert and send a Substack article to your Kindle:

```
go run main.go -url https://example.substack.com/p/article-name
```

Or simply:

```
go run main.go https://example.substack.com/p/article-name
```

### Specifying Output Format

By default, the application converts articles to EPUB format, which is the recommended format for sending to Kindle devices. You can specify a different output format using the `-format` flag:

```
go run main.go -url https://example.substack.com/p/article-name -format azw3
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
- Converts content to EPUB (default), AZW3, or MOBI format
- Direct conversion to AZW3 and MOBI formats without requiring Calibre
- Uses Calibre for conversion when available (better quality)
- Sends the converted file directly to your Kindle device
- Cleans up temporary files after sending

## How It Works

1. **Scraping**: Extracts article content, title, author, and images from the Substack URL
2. **Conversion**: 
   - For EPUB: Converts the content directly to EPUB format
   - For AZW3/MOBI: Converts directly to the requested format
   - If Calibre is available, it will be used for better quality conversion
3. **Delivery**: Sends the converted file to your Kindle email address

## Limitations

- Works only with public Substack articles (no paywall content)
- Some complex formatting or interactive elements may not be preserved
- MOBI format is no longer supported by Amazon's Send to Kindle service

## Project Structure

- `main.go`: Main application entry point
- `pkg/scraper`: Module for extracting content from Substack articles
- `pkg/converter`: Module for converting articles to EPUB, AZW3, or MOBI format
- `pkg/sender`: Module for sending files to Kindle via email 