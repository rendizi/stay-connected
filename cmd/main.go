package main 

import(
	"github.com/joho/godotenv"
	"log"
	"stay-connected/internal/services/inst"
	"os"
)

func main(){
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal(err)
	}
	inst, err := stories.LoginToInst(os.Getenv("TEST_INSTAGRAM_USERNAME"),os.Getenv("TEST_INSTAGRAM_PASSWORD"))
	if err != nil{
		log.Fatal(err)
	}
	l := stories.SummarizeInstagramStories(inst)
	log.Println(l[0])
}