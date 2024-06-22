package stories

import (
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"log"
	"stay-connected/internal/services/openai"
)

func LoginToInst(login string, password string) (*goinsta.Instagram, error) {
	insta := goinsta.New(login, password)
	err := insta.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to login to Instagram with username %s: %w", login, err)
	}
	return insta, nil
}

func SummarizeInstagramStories(insta *goinsta.Instagram)([]string){
	stories := insta.Timeline.Stories()
	storiesArray := make([]string, 0)

	for _, story := range stories {
		if story.User.Username != "armansu"{
			continue 
		}
		temp := make([]string,0)

		
		if story != nil {
			log.Println("By:",story.User.Username)
			profile, err := insta.VisitProfile(story.User.Username)
			if err != nil{
				log.Println("Error while getting "+story.User.Username+"' profile stories", err.Error)
				continue 
			}

			profilesStories, err := profile.User.Stories()
			if err != nil{
				log.Println("Error while getting "+story.User.Username+"' profile stories", err.Error)
				continue 
			}


			for _, stories := range profilesStories.Reel.Items{
				for _, media := range stories.Images.Versions{
					prompt := fmt.Sprintf("I have an image from an %s's(use it when want to write about him instead of writing 'user') Instagram story. Your task is to determine if it contains any interesting or relevant information about the person's life. If it does, summarize this information in 1 short sentence. If the image content is not related to the person's personal life, not inteserting or important activities, return empty response. Give logically connected summarize based on old storieses:%s. Don't repeat what is already summarized and in old storieses. Additional stories info: events: %s, hashtags: %s, polls: %s, locations: %s, questions: %s, sliders: %s, mentions: %s",
					 story.User.Username, temp, stories.StoryEvents, stories.StoryHashtags, stories.StoryPolls, stories.StoryLocations, stories.StorySliders, stories.StoryQuestions, stories.Mentions)
					resp,err := openai.SummarizeImage(media.URL,prompt)
					if err != nil{
						log.Println(err)
					}
					temp = append(temp, resp)
					break 
				}
			}
		}

		summarize, err := openai.SummarizeImagesToOne(temp)
		if err != nil{
			continue 
		}
		storiesArray = append(storiesArray, summarize)


	}
	return storiesArray
}