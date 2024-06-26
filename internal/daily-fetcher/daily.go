package daily

import (
	"fmt"
	"log"
	"os"
	"stay-connected/internal/encryption"
	"stay-connected/internal/services/db"
	"stay-connected/internal/services/inst"
	"stay-connected/internal/services/mailer"
	"stay-connected/internal/services/openai"
	"stay-connected/internal/services/telegram"
	"strings"
	"sync"
	"time"
)

func Do() {
	const maxRetries = 5
	var instas []db.Instagram
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		instas, err = db.GetInstas()
		if err == nil {
			processUsers(instas)
			return
		}
		log.Printf("Attempt %d/%d failed: %v", attempt, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	log.Printf("Failed to fetch users after %d attempts: %v", maxRetries, err)

	mailer.Send(os.Getenv("EMAIL_AUTHOR"), "Goroutine has reached maximum count of attemts to run")
}

func processUsers(users []db.Instagram) {
	secret := []byte(os.Getenv("SECRET_KEY_PASSWORD"))
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)

		go func(u db.Instagram) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			password, err := encryption.Decrypt(u.Password, secret)
			if err != nil {
				log.Println(err)
				return
			}

			inst, err := stories.LoginToInst(u.Username, password)
			if err != nil {
				log.Println(err)
				return
			}
			email, telegramId, err := db.GetEmail(u.UserId)
			if err != nil {
				log.Println(err)
				return
			}

			left, err := db.LeftToReactLimit(u.UserId)
			if err != nil {
				log.Println(err)
				return
			}
			if left <= 0 {
				return
			}

			used, url, result := stories.SummarizeInstagramStories(inst, left)

			if len(result) == 0 {
				return
			}

			if telegramId != -1 {
				if url != "" {
					err = telegram.SendAlbum(telegramId, url)
					if err != nil {
						log.Println(err)
						return
					}
				}
				formatted := formatStoriesForTelegram(result)
				err = telegram.SendMessage(telegramId, formatted)
				if err != nil {
					log.Println(err)
					return
				}
			}
			formatter := formatStoriesForEmail(result, url)
			err = mailer.Send(email, formatter)
			if err != nil {
				log.Println(err)
				return
			}

			err = db.Used(u.UserId, used)
			if err != nil {
				log.Println(err)
				return
			}
			if used >= left {
				if telegramId != -1 {
					telegram.SendMessage(telegramId, "Hey, you reached your usage limit. No more summarizes will be sent to you")
				}
				mailer.Send(email, "Hey, you reached your usage limit. No more summarizes will be sent to you")
			}
		}(user)
	}

	wg.Wait()
}

func formatStoriesForTelegram(stories []openai.StoriesType) string {
	var builder strings.Builder

	builder.WriteString("What is going on today?:\n\n")
	for _, story := range stories {
		builder.WriteString(fmt.Sprintf("• %s: %s\n", story.Author, story.Summarize))
	}

	return builder.String()
}

func formatStoriesForEmail(stories []openai.StoriesType, url string) string {
	var builder strings.Builder

	builder.WriteString(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: auto;
            background: #f9f9f9;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }
        .title {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 10px;
        }
		.video {
            margin-bottom: 20px;
            text-align: center;
        }
        .video iframe {
            width: 100%;
            max-width: 560px;
            height: 315px;
            border: none;
            border-radius: 10px;
            
        .story {
            margin-bottom: 20px;
        }
        .summary {
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="title">Daily Stories</div>
<div class="video">
            <iframe src="` + url + `" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
        </div>
        <hr>
`)

	for _, story := range stories {
		builder.WriteString(fmt.Sprintf(`
        <div class="story">
            <div class="summary"><span style="font-weight: bold; color: #555;">%s:</span>%s</div>
        </div>
        `, story.Author, story.Summarize))
	}

	builder.WriteString(`
    </div>
</body>
</html>
`)

	return builder.String()
}
