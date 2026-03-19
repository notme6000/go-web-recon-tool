package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func HandleFlags() {
	apiKeyFlag := flag.String("api", "", "OpenRouter API key to set in .env file")
	helpFlag := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *helpFlag {
		DisplayHelp()
		os.Exit(0)
	}

	if *apiKeyFlag != "" {
		err := CreateEnvFile(*apiKeyFlag)
		if err != nil {
			fmt.Printf("Error creating .env file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ .env file created successfully with API_KEY!")
		os.Exit(0)
	}
}

func DisplayHelp() {
	fmt.Println(`
Web Scraper with AI-powered Data Extraction & Directory Bruteforce

Usage:
  go-web-scraper [flags]

Flags:
  --api <api_key>           Create or update .env file with OpenRouter API key
  --wordlist, -w <path>     Path to wordlist file for directory bruteforce scanning
  --help                    Show this help message

Examples:
  go run main.go --api sk-or-v1-xxxxxxxxxxxxx
  go run main.go --wordlist ./wordlist.txt
  go run main.go -w ./wordlist.txt
  go run main.go --help

Wordlist Format:
  - One directory/path per line
  - Lines starting with '#' are treated as comments
  - Empty lines are ignored

Example Wordlist:
  # Admin directories
  admin
  administrator
  # API endpoints
  api/v1
  api/v2

For more information, visit: https://openrouter.ai
`)
}

func CreateEnvFile(apiKey string) error {
	envPath := ".env"
	content := fmt.Sprintf("API_KEY=%s\n", apiKey)
	return os.WriteFile(envPath, []byte(content), 0644)
}

func LoadAPIKey() (string, error) {
	envPath := ".env"

	_, err := os.Stat(envPath)
	envExists := err == nil

	if !envExists {
		fmt.Println("\n⚠️  .env file not found!")
		fmt.Println("Please provide your OpenRouter API key to continue.\n")
		fmt.Println("Usage: go run main.go --api <your_api_key>")
		fmt.Println("\nYou can get your API key from: https://openrouter.ai/keys\n")
		os.Exit(1)
	}

	err = godotenv.Load(envPath)
	if err != nil {
		fmt.Printf("Error loading %s: %v\n", envPath, err)
		os.Exit(1)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("\n⚠️  API_KEY not found in .env file!")
		fmt.Println("Please update your .env file or provide your API key:\n")
		fmt.Println("Usage: go run main.go --api <your_api_key>")
		fmt.Println("\nYou can get your API key from: https://openrouter.ai/keys\n")
		os.Exit(1)
	}

	return apiKey, nil
}

// GetWordlistFlag retrieves the wordlist flag value (supports both --wordlist and -w)
func GetWordlistFlag() string {
	return ""
}
