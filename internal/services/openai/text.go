package openai 

import(
	"github.com/go-resty/resty/v2"
	"encoding/json"
	"os"
	"fmt"
)

func SummarizeImagesToOne(userPrompt []string)(string,error){
	apiKey := os.Getenv("OPENAI_KEY")
	apiEndpoint := "https://api.openai.com/v1/chat/completions"

	client := resty.New()

	response, err := client.R().
	SetAuthToken(apiKey).
	SetHeader("Content-Type", "application/json").
	SetBody(map[string]interface{}{
		"model": "gpt-4o", 
		"messages": []interface{}{
			map[string]interface{}{
				"role": "system",
				"content": "You are given array of storieses summarize. I am very busy so give the most interesting ones, make them shorter without losing an idea. Maximum symbols-150, don't use markup symbols. Response should be like 1 text, no need to divide into ordered/unordered list",
			},
			map[string]interface{}{
				"role": "user",
				"content": fmt.Sprintf("%s",userPrompt),
			},
		},
		"max_tokens": 150,
	}).
	Post(apiEndpoint)


    if err != nil {
        return "",fmt.Errorf("Error while sending send the request: %v", err)
    }

    body := response.Body()

    var data map[string]interface{}
    err = json.Unmarshal(body, &data)
    if err != nil {
        return "", fmt.Errorf("Error while decoding JSON response:", err)
    }

    content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	return content, nil 
}