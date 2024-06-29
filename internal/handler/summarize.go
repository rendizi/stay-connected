package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	stories "stay-connected/internal/services/inst"
	"stay-connected/internal/services/openai"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections
		return true
	},
}

func Summarize(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting WebSocket connection upgrade...")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to set WebSocket upgrade: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	log.Println("WebSocket connection established.")
	defer func() {
		log.Println("Closing WebSocket connection...")
		conn.Close()
		log.Println("WebSocket connection closed.")
	}()

	for {
		// Step 1: Login to Instagram
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message from client:", err)
			sendMessage(conn, "Error reading message")
			break
		}
		log.Printf("Received message from client: %s", p)
		if sendMessageWithCheck(conn, "Read message from client") {
			break
		}

		log.Println("Attempting to login to Instagram...")
		if sendMessageWithCheck(conn, "Attempting to login to Instagram...") {
			break
		}

		inst, err := stories.LoginToInst(os.Getenv("TEST_INSTAGRAM_USERNAME"), os.Getenv("TEST_INSTAGRAM_PASSWORD"))
		if err != nil {
			log.Printf("Failed to login to Instagram: %v", err)
			sendMessage(conn, "Failed to login")
			break
		}
		log.Println("Successfully logged in to Instagram.")
		if sendMessageWithCheck(conn, "Successfully logged in to Instagram.") {
			break
		}

		// Step 3: Visit profile
		log.Printf("Attempting to visit Instagram profile: %s", string(p))
		if sendMessageWithCheck(conn, fmt.Sprintf("Attempting to visit Instagram profile: %s", string(p))) {
			break
		}
		profile, err := inst.VisitProfile(string(p))
		if err != nil {
			log.Printf("Failed to visit Instagram profile: %v", err)
			sendMessage(conn, "Failed to visit profile")
			break
		}
		log.Println("Successfully visited Instagram profile.")
		if sendMessageWithCheck(conn, "Visited profile") {
			break
		}

		// Step 4: Get user's stories
		log.Println("Attempting to retrieve user's stories...")
		if sendMessageWithCheck(conn, "Attempting to retrieve user's stories...") {
			break
		}
		profilesStories, err := profile.User.Stories()
		if err != nil {
			log.Printf("Failed to retrieve user's stories: %v", err)
			sendMessage(conn, "Failed to get user's stories")
			break
		}
		log.Println("Successfully retrieved user's stories.")
		if sendMessageWithCheck(conn, "Got user's stories") {
			break
		}

		// Step 5: Process stories and send summarized results
		temp := make([]openai.StoriesType, 0)
		log.Println("Processing user's stories for summarization...")
		if sendMessageWithCheck(conn, "Processing user's stories for summarization...") {
			break
		}
		for i, stories := range profilesStories.Reel.Items {
			if i >= 10 {
				break
			}
			for _, media := range stories.Images.Versions {
				log.Printf("Summarizing image: %s", media.URL)
				prompt := fmt.Sprintf("I have an image from an %s's(use it when want to write about him instead of writing 'user') Instagram story. Your task is to determine if it contains any interesting or relevant information about the person's life. If it does, summarize this information in 1 short sentence. If the image content is not related to the person's personal life, not inteserting or important activities, return following response: 'Nothing interesting'. Give logically connected summarize based on old storieses(if it is empty- don't say me it is empty, give result only based on photo or return empty response):%s. Don't repeat what is already summarized and in old storieses. Additional stories info: events: %s, hashtags: %s, polls: %s, locations: %s, questions: %s, sliders: %s, mentions: %s. Maximum tokens: 75, write it as simple as possible, like people would say, use simple words",
					stories.User.Username, temp, stories.StoryEvents, stories.StoryHashtags, stories.StoryPolls, stories.StoryLocations, stories.StorySliders, stories.StoryQuestions, stories.Mentions)
				resp, err := openai.SummarizeImage(media.URL, prompt)
				if err != nil {
					log.Printf("Failed to summarize image: %v", err)
					sendMessage(conn, "Failed to summarize image")
					break
				}
				log.Println("Image summarized successfully.")
				if sendMessageWithCheck(conn, "Summarized image") {
					break
				}

				if resp != "Nothing interesting" {
					tempStoriesType := openai.StoriesType{
						Author:    string(p),
						Summarize: resp,
					}
					if sendMessageWithCheck(conn, tempStoriesType.Summarize) {
						break
					}
					if profile.User.Friendship.FollowedBy {
						temp = append([]openai.StoriesType{tempStoriesType}, temp...)
					} else {
						temp = append(temp, tempStoriesType)
					}
				}
				break
			}
		}

		// Summarize all images to one result
		log.Println("Summarizing all processed images into one result...")
		if sendMessageWithCheck(conn, "Summarizing all processed images into one result...") {
			break
		}
		summarize, err := openai.SummarizeImagesToOne(temp)
		if err != nil {
			log.Printf("Failed to summarize images into one result: %v", err)
			sendMessage(conn, "Failed to summarize images to one")
			break
		}
		log.Println("Images summarized into one result successfully.")
		if sendMessageWithCheck(conn, summarize) {
			break
		}

		// Close the connection after processing
		log.Println("Finished processing. Closing connection.")
		conn.Close()
		break
	}
}

func sendMessageWithCheck(conn *websocket.Conn, message string) bool {
	log.Printf("Sending message to client: %s", message)
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error sending message '%s': %v", message, err)
		log.Println("Closing WebSocket connection due to send error...")
		conn.Close()
		log.Println("WebSocket connection closed due to send error.")
		return true
	}
	return false
}

func sendMessage(conn *websocket.Conn, message string) {
	log.Printf("Sending message to client: %s", message)
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error sending message '%s': %v", message, err)
		log.Println("Closing WebSocket connection due to send error...")
		conn.Close()
		log.Println("WebSocket connection closed due to send error.")
		return
	}
}
