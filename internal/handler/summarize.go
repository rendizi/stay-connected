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
		sendMessage(conn, "Read message from client")

		log.Println("Attempting to login to Instagram...")
		inst, err := stories.LoginToInst(os.Getenv("TEST_INSTAGRAM_USERNAME"), os.Getenv("TEST_INSTAGRAM_PASSWORD"))
		if err != nil {
			log.Printf("Failed to login to Instagram: %v", err)
			sendMessage(conn, "Failed to login")
			break
		}
		log.Println("Successfully logged in to Instagram.")
		sendMessage(conn, "Logged in to Instagram")

		// Step 2: Read message from client
		log.Println("Waiting to read message from client...")

		// Step 3: Visit profile
		log.Printf("Attempting to visit Instagram profile: %s", string(p))
		profile, err := inst.VisitProfile(string(p))
		if err != nil {
			log.Printf("Failed to visit Instagram profile: %v", err)
			sendMessage(conn, "Failed to visit profile")
			break
		}
		log.Println("Successfully visited Instagram profile.")
		sendMessage(conn, "Visited profile")

		// Step 4: Get user's stories
		log.Println("Attempting to retrieve user's stories...")
		profilesStories, err := profile.User.Stories()
		if err != nil {
			log.Printf("Failed to retrieve user's stories: %v", err)
			sendMessage(conn, "Failed to get user's stories")
			break
		}
		log.Println("Successfully retrieved user's stories.")
		sendMessage(conn, "Got user's stories")

		// Step 5: Process stories and send summarized results
		temp := make([]openai.StoriesType, 0)
		log.Println("Processing user's stories for summarization...")
		for _, story := range profilesStories.Reel.Items {
			for _, media := range story.Images.Versions {
				log.Printf("Summarizing image: %s", media.URL)
				prompt := fmt.Sprintf("Summarize this story with context: %s", string(p))
				resp, err := openai.SummarizeImage(media.URL, prompt)
				if err != nil {
					log.Printf("Failed to summarize image: %v", err)
					sendMessage(conn, "Failed to summarize image")
					continue
				}
				log.Println("Image summarized successfully.")
				sendMessage(conn, "Summarized image")

				if resp != "Nothing interesting" {
					tempStoriesType := openai.StoriesType{
						Author:    string(p),
						Summarize: resp,
					}
					if profile.User.Friendship.FollowedBy {
						temp = append([]openai.StoriesType{tempStoriesType}, temp...)
					} else {
						temp = append(temp, tempStoriesType)
					}
				}
			}
		}

		// Summarize all images to one result
		log.Println("Summarizing all processed images into one result...")
		summarize, err := openai.SummarizeImagesToOne(temp)
		if err != nil {
			log.Printf("Failed to summarize images into one result: %v", err)
			sendMessage(conn, "Failed to summarize images to one")
			break
		}
		log.Println("Images summarized into one result successfully.")
		sendMessage(conn, summarize)

		// Close the connection after processing
		log.Println("Finished processing. Closing connection.")
		conn.Close()
		break
	}
}

func sendMessage(conn *websocket.Conn, message string) {
	log.Printf("Sending message to client: %s", message)
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error sending message '%s': %v", message, err)
	}
}
