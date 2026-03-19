package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/notme6000/go-scrape/pkg/api"
	"github.com/notme6000/go-scrape/pkg/bruteforce"
	"github.com/notme6000/go-scrape/pkg/cli"
	"github.com/notme6000/go-scrape/pkg/extractor"
	"github.com/notme6000/go-scrape/pkg/scraper"
	"github.com/notme6000/go-scrape/pkg/types"
)

func printExtractedData(data types.ExtractedData) {
	fmt.Println("\nExtracted Data:")
	fmt.Println()
	
	fmt.Println("NAMES")
	fmt.Println("-----")
	if len(data.Names) > 0 {
		for _, name := range data.Names {
			fmt.Println(name)
		}
	} else {
		fmt.Println("No names found")
	}
	fmt.Println()
	
	fmt.Println("EMAILS")
	fmt.Println("------")
	if len(data.Emails) > 0 {
		for _, email := range data.Emails {
			fmt.Println(email)
		}
	} else {
		fmt.Println("No emails found")
	}
	fmt.Println()
	
	fmt.Println("PHONE NUMBERS")
	fmt.Println("-------------")
	if len(data.PhoneNumbers) > 0 {
		for _, phone := range data.PhoneNumbers {
			fmt.Println(phone)
		}
	} else {
		fmt.Println("No phone numbers found")
	}
	fmt.Println()
	
	fmt.Println("ADDRESSES")
	fmt.Println("---------")
	if len(data.Addresses) > 0 {
		for _, address := range data.Addresses {
			fmt.Println(address)
		}
	} else {
		fmt.Println("No addresses found")
	}
	fmt.Println()
}

func scrapeDirectoryData(url string, browser *scraper.Browser, apiClient *api.Client) (types.ExtractedData, error) {
	body, err := browser.ScrapeWebsite(url)
	if err != nil {
		return types.ExtractedData{}, err
	}

	namesCandidates := extractor.ExtractNames(body)
	validatedNames := apiClient.ValidateNamesWithAI(namesCandidates)

	return types.ExtractedData{
		Emails:       extractor.ExtractEmails(body),
		Names:        validatedNames,
		PhoneNumbers: extractor.ExtractPhoneNumbers(body),
		Addresses:    extractor.ExtractAddresses(body),
	}, nil
}

func printDirectoryData(path string, data types.ExtractedData) {
	fmt.Printf("\n📁 Directory: %s\n", path)
	fmt.Println(strings.Repeat("-", 50))
	
	if len(data.Names) > 0 {
		fmt.Println("NAMES:")
		for _, name := range data.Names {
			fmt.Printf("  • %s\n", name)
		}
	}
	
	if len(data.Emails) > 0 {
		fmt.Println("EMAILS:")
		for _, email := range data.Emails {
			fmt.Printf("  • %s\n", email)
		}
	}
	
	if len(data.PhoneNumbers) > 0 {
		fmt.Println("PHONE NUMBERS:")
		for _, phone := range data.PhoneNumbers {
			fmt.Printf("  • %s\n", phone)
		}
	}
	
	if len(data.Addresses) > 0 {
		fmt.Println("ADDRESSES:")
		for _, address := range data.Addresses {
			fmt.Printf("  • %s\n", address)
		}
	}
	
	if len(data.Names) == 0 && len(data.Emails) == 0 && len(data.PhoneNumbers) == 0 && len(data.Addresses) == 0 {
		fmt.Println("No data found in this directory")
	}
}

func main() {
	// Parse wordlist flag before HandleFlags
	wordlistFlag := flag.String("wordlist", "", "Path to wordlist file for directory bruteforce scanning")
	wordlistFlagShort := flag.String("w", "", "Shorthand for --wordlist")
	flag.Parse()
	
	wordlistPath := *wordlistFlag
	if wordlistPath == "" {
		wordlistPath = *wordlistFlagShort
	}

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

	apiClient := api.NewClient(apiKey)

	// Directory bruteforce scanning - FIRST
	var allDirectoryData []map[string]interface{}
	
	if wordlistPath != "" {
		fmt.Println("\n--- Starting Directory Bruteforce Scan ---")
		scanner := bruteforce.NewBruteforceScanner(website, 10, 5*time.Second)
		
		fmt.Printf("Loading wordlist from: %s\n", wordlistPath)
		err := scanner.LoadWordlistFromFile(wordlistPath)
		if err != nil {
			fmt.Printf("⚠️  Wordlist error: %v\n", err)
			return
		}

		fmt.Printf("Loaded %d paths from wordlist. Starting scan...\n\n", len(scanner.Wordlist))
		results := scanner.Scan()
		bruteforce.Print(results)

		if len(results) > 0 {
			fmt.Printf("\n\n--- Scraping Found Directories ---\n")
			fmt.Printf("Found %d directories. Starting data extraction...\n\n", len(results))

			// Scrape each found directory concurrently
			directoryChan := make(chan map[string]interface{}, len(results))
			var wg sync.WaitGroup

			for _, result := range results {
				wg.Add(1)
				go func(res bruteforce.DirectoryResult) {
					defer wg.Done()

					fullURL := website + "/" + strings.TrimPrefix(res.Path, "/")
					fmt.Printf("Scraping: %s\n", fullURL)
					
					data, err := scrapeDirectoryData(fullURL, browser, apiClient)
					if err != nil {
						directoryChan <- map[string]interface{}{
							"path":        res.Path,
							"status_code": res.StatusCode,
							"error":       err.Error(),
						}
					} else {
						directoryChan <- map[string]interface{}{
							"path":        res.Path,
							"status_code": res.StatusCode,
							"data": map[string]interface{}{
								"emails":        data.Emails,
								"names":         data.Names,
								"phone_numbers": data.PhoneNumbers,
								"addresses":     data.Addresses,
							},
						}
					}
				}(result)
			}

			go func() {
				wg.Wait()
				close(directoryChan)
			}()

			// Collect and display results
			for result := range directoryChan {
				if errMsg, exists := result["error"]; exists {
					fmt.Printf("⚠️  Failed to scrape %s: %s\n", result["path"], errMsg)
				} else {
					// Extract data for display
					path := result["path"].(string)
					dataMap := result["data"].(map[string]interface{})
					data := types.ExtractedData{
						Emails:       convertToStringSlice(dataMap["emails"]),
						Names:        convertToStringSlice(dataMap["names"]),
						PhoneNumbers: convertToStringSlice(dataMap["phone_numbers"]),
						Addresses:    convertToStringSlice(dataMap["addresses"]),
					}
					printDirectoryData(path, data)
				}
				allDirectoryData = append(allDirectoryData, result)
			}

			// Save directory results to JSON
			comprehensiveData := map[string]interface{}{
				"base_url":    website,
				"directories": allDirectoryData,
				"total_dirs":  len(results),
			}
			comprehensiveJSON, err := json.MarshalIndent(comprehensiveData, "", "  ")
			if err == nil {
				os.WriteFile("data/complete_results.json", comprehensiveJSON, 0644)
				fmt.Printf("\n✓ Directory scan results saved to: data/complete_results.json\n")
			}
		}
	}

	// Base URL scraping - SECOND
	fmt.Println("\n--- Scraping Base URL ---")
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
	if len(allDirectoryData) > 0 {
		fmt.Println("  - data/complete_results.json (directory scan results)")
	}
	
	printExtractedData(extractedData)
}

// convertToStringSlice converts interface{} to []string
func convertToStringSlice(i interface{}) []string {
	if slice, ok := i.([]interface{}); ok {
		result := make([]string, 0, len(slice))
		for _, v := range slice {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}

