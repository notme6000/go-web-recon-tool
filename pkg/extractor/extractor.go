package extractor

import (
	"regexp"
	"strings"
)

func ExtractEmails(text string) []string {
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return emailPattern.FindAllString(text, -1)
}

func ExtractPhoneNumbers(text string) []string {
	phonePattern := regexp.MustCompile(`\+?1?\s?(\d{3}[-.\s]?)?\d{3}[-.\s]?\d{4}|(\d{10})`)
	return phonePattern.FindAllString(text, -1)
}

func ExtractAddresses(text string) []string {
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

func ExtractNames(text string) []string {
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
