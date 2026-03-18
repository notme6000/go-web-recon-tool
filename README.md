# Go Web Scraper

A Go-based web scraper that extracts contact information and personal data from websites using headless browser automation and AI-powered validation.

## Features

- **Headless Browser Scraping**: Uses [Rod](https://github.com/go-rod/rod) for automated web page rendering and text extraction
- **Email Extraction**: Identifies and extracts email addresses using regex pattern matching
- **Name Extraction**: Extracts potential person names with pattern matching and heuristics
- **AI-Powered Name Validation**: Validates extracted names using OpenRouter API with AI models to filter false positives
- **Phone Number Detection**: Extracts phone numbers in various formats (with/without country codes, different separators)
- **Address Extraction**: Identifies street addresses with structured patterns
- **Structured Output**: Saves results in both raw text and JSON formats

## Requirements

- Go 1.25.0 or higher
- Environment variable `API_KEY` with OpenRouter API key for AI-powered name validation

## Dependencies

- `github.com/go-rod/rod` (v0.116.2) - Headless browser automation
- `github.com/joho/godotenv` (v1.5.1) - Environment variable management

## Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd go-web-scraper-copiolet
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file in the root directory with your OpenRouter API key:
   ```
   API_KEY=your_openrouter_api_key_here
   ```

## Usage

Run the scraper from the command line:

```bash
go run main.go
```

The program will prompt you to enter a website URL:

```
enter the link: https://example.com
scanning https://example.com
```

The scraper will then:
1. Navigate to the specified website using a headless Chrome browser
2. Extract all text content
3. Parse for emails, phone numbers, and addresses using regex patterns
4. Extract name candidates and validate them using AI
5. Save results to the `data/` directory

## Output

The scraper creates two files in the `data/` directory:

### `data/data.txt`
Raw text content extracted from the website

### `data/extracted_data.json`
Structured JSON containing:
```json
{
  "emails": ["user@example.com"],
  "names": ["John Doe", "Jane Smith"],
  "phone_numbers": ["+1 (555) 123-4567", "555.987.6543"],
  "addresses": ["123 Main St", "456 Oak Avenue"]
}
```

## How It Works

### Name Extraction & Validation

The scraper uses a multi-stage approach for name detection:

1. **Candidate Extraction**: Identifies text lines matching typical name patterns:
   - Starts with capital letters
   - Contains 2-3 words
   - Between 5-50 characters
   - No special characters or numbers

2. **Filtering**: Removes common false positives:
   - Common words (articles, prepositions, etc.)
   - Lines ending in 's' (likely plurals)
   - Lines with invalid character combinations

3. **AI Validation**: Uses OpenRouter API to validate candidates with language models, confirming they are actual person names

### Contact Information Extraction

- **Emails**: RFC-compliant email pattern matching
- **Phone Numbers**: Supports multiple formats including international, dash/dot/space separators
- **Addresses**: Identifies lines with street number and common street suffixes (St, Ave, Blvd, etc.)

## Configuration

The main extraction parameters can be adjusted in `main.go`:

- Email regex pattern (line 43)
- Phone number regex pattern (line 125)
- Address pattern and validation (line 132)
- Name pattern rules (lines 147-170)
- Character length constraints for names (line 177)

## Error Handling

- Falls back to unvalidated candidates if AI validation fails
- Gracefully handles missing API key or connection issues
- Creates `data/` directory automatically if it doesn't exist

## License

MIT

## Author

[notme6000](https://github.com/notme6000)
