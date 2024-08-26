package function

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func init() {
	functions.HTTP("HelloAI", HelloAI)
}

func HelloAI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	fmt.Println("request method and path:", r.Method, r.URL.Path)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost && r.URL.Path != "/chat" {
		http.Error(w, "Only POST requests to /chat path are accepted", http.StatusMethodNotAllowed)
		return

	}

	ctx := r.Context()

	client, err := genai.NewClient(ctx, option.WithAPIKey(""))

	if err != nil {
		log.Printf("Error processing request: %v", err)
		http.Error(w, "Error on create gemini client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	//  gemini promodel
	model := client.GenerativeModel("gemini-pro")

	// read input from json body with key "input"
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Error decoding input", http.StatusBadRequest)
		return
	}

	inputText, ok := input["input"].(string)

	if !ok {
		http.Error(w, "Error reading input", http.StatusBadRequest)
		return
	}

	resp, err := model.GenerateContent(ctx, genai.Text(inputText))

	if err != nil {
		log.Printf("Error generating content: %v", err)
		http.Error(w, "Error generating content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": formatResponse(resp)})
}

// format resposne
func formatResponse(resp *genai.GenerateContentResponse) string {
	var formattedContent strings.Builder
	if resp != nil && resp.Candidates != nil {
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					formattedContent.WriteString(fmt.Sprintf("%v", part))
				}
			}
		}
	}

	return formattedContent.String()
}
