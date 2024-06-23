package main 

import(
	"github.com/joho/godotenv"
	"log"
	"stay-connected/internal/daily-fetcher"
    "github.com/jasonlvhit/gocron"

)

func main(){
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal(err)
	}

	

	//users, err := db.Insert(os.Getenv("TEST_INSTAGRAM_USERNAME"),os.Getenv("TEST_INSTAGRAM_PASSWORD"), "baglanov.a0930@gmail.com")
	//if err != nil{
	//	log.Fatal(err)
	//}
	//log.Println(users)

	daily.Do()

	s := gocron.NewScheduler()
    s.Every(1).Days().Do(daily.Do)
    <- s.Start()

	//fullusers, err := db.Get()
	//if err != nil{
	//	log.Fatal(err)
	//}
	//log.Println(fullusers)

	//return 

	//l := stories.SummarizeInstagramStories(inst)
	//log.Println(l[0])
}