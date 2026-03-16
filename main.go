package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"github.com/joho/godotenv"
)

type ExtractedData struct {
	Emails    []string `json:"emails"`
	Names     []string `json:"names"`
	PhoneNumbers []string `json:"phone_numbers"`
	Addresses []string `json:"addresses"`
}

type OpenRouterRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func extractEmails(text string) []string {
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return emailPattern.FindAllString(text, -1)
}

func validateNamesWithAI(candidates []string, apiKey string) []string {
	if len(candidates) == 0 {
		return []string{}
	}
	
	candidatesStr := strings.Join(candidates, ", ")
	
	prompt := fmt.Sprintf(`From the following list of text candidates, identify which ones are actual person names. Return ONLY a comma-separated list of valid names, nothing else.

Candidates: %s

Valid names:`, candidatesStr)
	
	reqBody := OpenRouterRequest{
		Model: "openrouter/auto",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error marshaling request:", err)
		return candidates
	}
	
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return candidates
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making API call:", err)
		return candidates
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return candidates
	}
	
	var response OpenRouterResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return candidates
	}
	
	if len(response.Choices) == 0 {
		return candidates
	}
	
	content := strings.TrimSpace(response.Choices[0].Message.Content)
	validNames := strings.Split(content, ",")
	
	var cleanNames []string
	for _, name := range validNames {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			cleanNames = append(cleanNames, trimmed)
		}
	}
	
	return cleanNames
}

func extractPhoneNumbers(text string) []string {
	phonePattern := regexp.MustCompile(`\+?1?\s?(\d{3}[-.\s]?)?\d{3}[-.\s]?\d{4}|(\d{10})`)
	return phonePattern.FindAllString(text, -1)
}

func extractAddresses(text string) []string {
	lines := strings.Split(text, "\n")
	var addresses []string
	addressPattern := regexp.MustCompile(`\d+\s+[a-zA-Z\s,]+(?:St|Street|Ave|Avenue|Blvd|Boulevard|Rd|Road|Dr|Drive|Ln|Lane|Ct|Court|Pl|Place|Way)`)
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 10 && addressPattern.MatchString(trimmed) {
			addresses = append(addresses, trimmed)
		}
	}
	return addresses
}

func extractNames(text string) []string {
	lines := strings.Split(text, "\n")
	var candidates []string
	
	excludeWords := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "from": true,
		"about": true, "contact": true, "email": true, "phone": true,
		"address": true, "website": true, "home": true, "office": true,
		"business": true, "company": true, "service": true, "product": true,
		"click": true, "here": true, "read": true, "more": true, "view": true,
		"all": true, "this": true, "that": true, "which": true, "who": true,
		"welcome": true, "hello": true, "thanks": true, "please": true,
		"price": true, "cost": true, "date": true, "time": true, "number": true,
		"name": true, "type": true, "detail": true, "details": true, "info": true,
		"information": true, "description": true, "title": true, "subject": true,
		"best": true, "new": true, "top": true, "latest": true, "featured": true,
		"popular": true, "special": true, "offer": true, "sale": true, "free": true,
		"call": true, "visit": true, "follow": true, "like": true, "share": true,
		"subscribe": true, "download": true, "upload": true, "login": true, "register": true,
		"search": true, "find": true, "buy": true, "sell": true, "shop": true,
		"section": true, "page": true, "menu": true, "link": true, "button": true,
		"list": true, "category": true, "tag": true, "label": true, "code": true,
		"location": true, "place": true, "city": true, "state": true, "country": true,
		"street": true, "avenue": true, "boulevard": true, "road": true, "drive": true,
		"results": true, "content": true, "text": true, "image": true, "video": true,
		"article": true, "blog": true, "post": true, "comment": true, "reply": true,
		"me": true, "my": true, "your": true, "project": true,
	}
	
	namePattern := regexp.MustCompile(`^[A-Z][a-z]+(?:\s+[A-Z][a-z]+){1,2}$`)
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if len(trimmed) < 5 || len(trimmed) > 50 {
			continue
		}
		
		if strings.ContainsAny(trimmed, "0123456789@#$%^&*()-_+=[]{}|;:',.<>?/~`\\") {
			continue
		}
		
		if !namePattern.MatchString(trimmed) {
			continue
		}
		
		words := strings.Fields(trimmed)
		
		if len(words) < 2 || len(words) > 3 {
			continue
		}
		
		validLength := true
		for _, word := range words {
			if len(word) < 2 {
				validLength = false
				break
			}
		}
		if !validLength {
			continue
		}
		
		isExcluded := false
		for _, word := range words {
			if excludeWords[strings.ToLower(word)] {
				isExcluded = true
				break
			}
		}
		
		if isExcluded {
			continue
		}
		
		allPlural := true
		for _, word := range words {
			if !strings.HasSuffix(strings.ToLower(word), "s") {
				allPlural = false
				break
			}
		}
		if allPlural {
			continue
		}
		
		candidates = append(candidates, trimmed)
	}
	
	return candidates
}


func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	var website string
	apiKey := os.Getenv("API_KEY")

	fmt.Printf("enter the link: ")
	fmt.Scan(&website)
	fmt.Println("scanning", website)

	err = os.MkdirAll("data", 0755)
	if err != nil {
		panic(err)
	}

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(website)
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36",
	})
	page.MustWaitLoad()

	body := page.MustElement("body").MustText()

	err = os.WriteFile("data/data.txt", []byte(body), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Extracting data...")
	
	namesCandidates := extractNames(body)
	fmt.Printf("Found %d name candidates, validating with AI...\n", len(namesCandidates))
	
	validatedNames := validateNamesWithAI(namesCandidates, apiKey)
	
	extractedData := ExtractedData{
		Emails:       extractEmails(body),
		Names:        validatedNames,
		PhoneNumbers: extractPhoneNumbers(body),
		Addresses:    extractAddresses(body),
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

