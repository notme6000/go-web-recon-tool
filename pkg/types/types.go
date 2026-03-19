package types

type ExtractedData struct {
	Emails       []string `json:"emails"`
	Names        []string `json:"names"`
	PhoneNumbers []string `json:"phone_numbers"`
	Addresses    []string `json:"addresses"`
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
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
