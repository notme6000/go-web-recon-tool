package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/notme6000/go-scrape/pkg/api"
	"github.com/notme6000/go-scrape/pkg/cli"
	"github.com/notme6000/go-scrape/pkg/extractor"
	"github.com/notme6000/go-scrape/pkg/scraper"
	"github.com/notme6000/go-scrape/pkg/types"
)

func main() {
	cli.HandleFlags()
	apiKey, _ := cli.LoadAPIKey()

	var website string

	fmt.Printf("enter the link: ")
	fmt.Scan(&website)
	fmt.Println("scanning", website)

	err := os.MkdirAll("data", 0755)
	if err != nil {
		panic(err)
	}

	browser, err := scraper.NewBrowser()
	if err != nil {
		panic(err)
	}
	defer browser.Close()

	body, err := browser.ScrapeWebsite(website)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("data/data.txt", []byte(body), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Extracting data...")

	namesCandidates := extractor.ExtractNames(body)
	fmt.Printf("Found %d name candidates, validating with AI...\n", len(namesCandidates))

	apiClient := api.NewClient(apiKey)
	validatedNames := apiClient.ValidateNamesWithAI(namesCandidates)

	extractedData := types.ExtractedData{
		Emails:       extractor.ExtractEmails(body),
		Names:        validatedNames,
		PhoneNumbers: extractor.ExtractPhoneNumbers(body),
		Addresses:    extractor.ExtractAddresses(body),
	}

	jsonData, err := json.MarshalIndent(extractedData, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("data/extracted_data.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Data extracted successfully!")
	fmt.Println("Files created in 'data' folder:")
	fmt.Println("  - data/data.txt (raw text)")
	fmt.Println("  - data/extracted_data.json (structured data)")
	fmt.Printf("\nExtracted Data:\n%s\n", string(jsonData))
}

