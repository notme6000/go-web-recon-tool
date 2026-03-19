package bruteforce

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type DirectoryResult struct {
	Path       string
	StatusCode int
	Exists     bool
}

type BruteforceScanner struct {
	BaseURL    string
	Client     *http.Client
	Wordlist   []string
	Timeout    time.Duration
	Threads    int
}

// NewBruteforceScanner creates a new scanner instance
func NewBruteforceScanner(baseURL string, threads int, timeout time.Duration) *BruteforceScanner {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	return &BruteforceScanner{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: timeout,
		},
		Wordlist: []string{},
		Timeout:  timeout,
		Threads:  threads,
	}
}

// LoadWordlistFromFile loads wordlist from an external file
func (bs *BruteforceScanner) LoadWordlistFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open wordlist file: %w", err)
	}
	defer file.Close()

	var wordlist []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			wordlist = append(wordlist, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading wordlist file: %w", err)
	}

	if len(wordlist) == 0 {
		return fmt.Errorf("wordlist file is empty or contains only comments")
	}

	bs.Wordlist = wordlist
	return nil
}

// Scan performs the directory bruteforce scan
func (bs *BruteforceScanner) Scan() []DirectoryResult {
	var results []DirectoryResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create a channel for distributing work
	paths := make(chan string, bs.Threads)
	
	// Start worker goroutines
	for i := 0; i < bs.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				result := bs.checkPath(path)
				if result.Exists {
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
				}
			}
		}()
	}

	// Send paths to workers
	go func() {
		for _, path := range bs.Wordlist {
			paths <- path
		}
		close(paths)
	}()

	wg.Wait()
	return results
}

// checkPath checks if a single path exists on the server
func (bs *BruteforceScanner) checkPath(path string) DirectoryResult {
	url := bs.BaseURL + "/" + strings.TrimPrefix(path, "/")
	
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return DirectoryResult{
			Path:       path,
			StatusCode: 0,
			Exists:     false,
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	resp, err := bs.Client.Do(req)
	if err != nil {
		// Try GET if HEAD fails
		req.Method = "GET"
		resp, err = bs.Client.Do(req)
		if err != nil {
			return DirectoryResult{
				Path:       path,
				StatusCode: 0,
				Exists:     false,
			}
		}
	}
	defer resp.Body.Close()

	// Consider 200-299 and some 3xx codes as existing
	exists := (resp.StatusCode >= 200 && resp.StatusCode < 400) || 
		resp.StatusCode == 403 || resp.StatusCode == 405

	return DirectoryResult{
		Path:       path,
		StatusCode: resp.StatusCode,
		Exists:     exists,
	}
}

// Print prints the scan results in readable format
func Print(results []DirectoryResult) {
	if len(results) == 0 {
		fmt.Println("No directories found")
		return
	}

	fmt.Println("\nDIRECTORIES FOUND")
	fmt.Println("----------------")
	for _, result := range results {
		statusStr := getStatusString(result.StatusCode)
		fmt.Printf("%-30s [%d] %s\n", result.Path, result.StatusCode, statusStr)
	}
}

// getStatusString returns human-readable status description
func getStatusString(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "OK"
	case code >= 300 && code < 400:
		return "Redirect"
	case code == 403:
		return "Forbidden"
	case code == 404:
		return "Not Found"
	case code == 405:
		return "Method Not Allowed"
	default:
		return "Unknown"
	}
}
