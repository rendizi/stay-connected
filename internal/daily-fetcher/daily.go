package daily 

import(
	"stay-connected/internal/services/db"
	"log"
	"time"
	"stay-connected/internal/encryption"
	"os"
	"stay-connected/internal/services/inst"
	"sync"
)

func Do(){
	const maxRetries = 5 
	var users []db.FullUser
	var err error 
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		users, err = db.Get()
		if err == nil{
			processUsers(users)
			return 
		}
		log.Printf("Attempt %d/%d failed: %v", attempt, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	log.Printf("Failed to fetch users after %d attempts: %v", maxRetries, err)

	//email me
}

func processUsers(users []db.FullUser) {
	secret := []byte(os.Getenv("SECRET_KEY_PASSWORD"))
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)

		go func(u db.FullUser) {
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

			result := stories.SummarizeInstagramStories(inst)
			log.Println(result)
		}(user)
	}

	wg.Wait()
}
