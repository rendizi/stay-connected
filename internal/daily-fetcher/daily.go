package daily

import (
	"fmt"
	"log"
	"os"
	"stay-connected/internal/encryption"
	"stay-connected/internal/services/db"
	"stay-connected/internal/services/inst"
	"stay-connected/internal/services/mailer"
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

			left, err := db.LeftToReactLimit(u.UserId)
			if err != nil {
				log.Println(err)
				return
			}

			used, result := stories.SummarizeInstagramStories(inst, left)

			for _, res := range result {
				log.Println(res)
			}

			email, err := db.GetEmail(u.UserId)
			if err != nil {
				log.Println(err)
				return
			}
			//send via email
			err = mailer.Send(email, fmt.Sprintf("%s", result))
			if err != nil {
				log.Println(err)
				return
			}

			err = db.Used(u.UserId, used)
			if err != nil {
				log.Println(err)
				return
			}
		}(user)
	}

	wg.Wait()
}
