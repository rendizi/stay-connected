package openai

import(
	"github.com/go-resty/resty/v2"
	"encoding/json"
	"os"
	"fmt"
)

func SummarizeImage(url string, prompt string)(string, error){
	apiEndpoint := "https://api.openai.com/v1/chat/completions"

	apiKey := os.Getenv("OPENAI_KEY")
    client := resty.New()

    response, err := client.R().
        SetAuthToken(apiKey).
        SetHeader("Content-Type", "application/json").
        SetBody(map[string]interface{}{
			"model": "gpt-4o",
			"messages": []interface{}{
				map[string]interface{}{
					"role": "user",
					"content": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": prompt,
						},
						map[string]interface{}{
							"type": "image_url",
							"image_url": map[string]interface{}{
								"url": url,
							},
						},
					},
				},
			},
			"max_tokens": 75,
		}).
        Post(apiEndpoint)

    if err != nil {
        return "",fmt.Errorf("Error while sending send the request: %v", err)
    }

    body := response.Body()

    var data map[string]interface{}
    err = json.Unmarshal(body, &data)
    if err != nil {
        return "",fmt.Errorf("Error while decoding JSON response:", err)
    }

    content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
    return content, nil 
}