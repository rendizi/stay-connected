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
	log.Println("websocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	for {
		log.Println(32)
		// Step 1: Login to Instagram
		inst, err := stories.LoginToInst(os.Getenv("TEST_INSTAGRAM_USERNAME"), os.Getenv("TEST_INSTAGRAM_PASSWORD"))
		if err != nil {
			log.Printf("Failed to login: %v", err)
			sendMessage(conn, "Failed to login")
			return
		}
		sendMessage(conn, "Logged in to Instagram")

		// Step 2: Read message from client
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			sendMessage(conn, "Error reading message")
			return
		}
		log.Printf("Received message: %s\n", p)
		sendMessage(conn, "Read message from client")

		// Step 3: Visit profile
		profile, err := inst.VisitProfile(string(p))
		if err != nil {
			log.Printf("Failed to visit profile: %v", err)
			sendMessage(conn, "Failed to visit profile")
			return
		}
		sendMessage(conn, "Visited profile")

		// Step 4: Get user's stories
		profilesStories, err := profile.User.Stories()
		if err != nil {
			log.Printf("Failed to get user's stories: %v", err)
			sendMessage(conn, "Failed to get user's stories")
			return
		}
		sendMessage(conn, "Got user's stories")

		// Step 5: Process stories and send summarized results
		temp := make([]openai.StoriesType, 0)
		for _, stories := range profilesStories.Reel.Items {
			for _, media := range stories.Images.Versions {
				// Summarize image
				prompt := fmt.Sprintf("Prompt for summarization with %s", string(p))
				resp, err := openai.SummarizeImage(media.URL, prompt)
				if err != nil {
					log.Printf("Failed to summarize image: %v", err)
					sendMessage(conn, "Failed to summarize image")
					continue
				}
				sendMessage(conn, "Summarized image")

				// Handle summarized result
				if resp != "Nothing interesting" {
					var tempStoriesType openai.StoriesType
					tempStoriesType.Author = string(p)
					tempStoriesType.Summarize = resp
					if profile.User.Friendship.FollowedBy {
						temp = append([]openai.StoriesType{tempStoriesType}, temp...)
					} else {
						temp = append(temp, tempStoriesType)
					}
				}
			}
		}

		// Summarize all images to one result
		summarize, err := openai.SummarizeImagesToOne(temp)
		if err != nil {
			log.Printf("Failed to summarize images to one: %v", err)
			sendMessage(conn, "Failed to summarize images to one")
			continue
		}
		sendMessage(conn, summarize)
		conn.Close()
	}
}

func sendMessage(conn *websocket.Conn, message string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error sending message '%s': %v", message, err)
	}
}
