package stories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Davincible/goinsta"
	"log"
	"math/rand"
	"stay-connected/internal/services/openai"
	"stay-connected/internal/services/redis"
	"time"
)

func LoginToInst(login string, password string) (*goinsta.Instagram, error) {
	insta := goinsta.New(login, password)
	err := insta.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to login to Instagram with username %s: %w", login, err)
	}
	return insta, nil
}

func SummarizeInstagramStories(insta *goinsta.Instagram, left int8) (int8, string, []openai.StoriesType) {
	stories := insta.Timeline.Stories()
	storiesArray := make([]openai.StoriesType, 0)
	var used int8
	used = 0
	reachedLimit := false
	log.Println("there")
	rand.Seed(time.Now().UnixNano())

	medias := make([]Asset, 0)

	for _, story := range stories {
		if reachedLimit {
			break
		}
		temp := make([]openai.StoriesType, 0)

		if story != nil {
			log.Println("By:", story.User.Username)
			profile, err := insta.VisitProfile(story.User.Username)
			if err != nil {
				log.Println("Error while getting "+story.User.Username+"' profile stories", err.Error)
				continue
			}

			profilesStories, err := profile.User.Stories()
			if err != nil {
				log.Println("Error while getting "+story.User.Username+"' profile stories", err.Error)
				continue
			}

			data, err := redis.GetFromRedis(context.Background(), story.User.Username)
			if err != nil {
				log.Println("Error while getting "+story.User.Username+"' stories summarize history", err.Error)
			}
			var thisWeek []string
			err = json.Unmarshal([]byte(data), &thisWeek)
			if err != nil {
				log.Println(err)
			}
			for _, stories := range profilesStories.Reel.Items {
				for _, media := range stories.Images.Versions {
					val, err := redis.GetFromRedis(context.Background(), media.URL)
					var resp string
					if err != nil {
						prompt := fmt.Sprintf("I have an image from an %s's(use it when want to write about him instead of writing 'user') Instagram story. Your task is to determine if it contains any interesting or relevant information about the person's life. If it does, summarize this information in 1 short sentence. If the image content is not related to the person's personal life, not inteserting or important activities, return following response: 'Nothing interesting'. Give logically connected summarize based on previous storieses(if it is empty- don't say me it is empty, give result only based on photo or return empty response):%s.Last 7 days stories: %s. Don't repeat what is already summarized and in old storieses. Additional stories info: events: %s, hashtags: %s, polls: %s, locations: %s, questions: %s, sliders: %s, mentions: %s. Maximum tokens: 75, write it as simple as possible, like people would say, use simple words",
							story.User.Username, temp, data, stories.StoryEvents, stories.StoryHashtags, stories.StoryPolls, stories.StoryLocations, stories.StorySliders, stories.StoryQuestions, stories.Mentions)
						resp, err = openai.SummarizeImage(media.URL, prompt)
						if err != nil {
							log.Println(err)
						}
						err = redis.StoreInRedis(context.Background(), media.URL, resp, 24*time.Hour)
						if err != nil {
							log.Println(err)
						}
						used += 1
						if used >= left {
							reachedLimit = true
						}
					} else {
						resp = val
					}

					if resp != "Nothing interesting" || resp != "Nothing interesting." {
						var tempStoriesType openai.StoriesType
						tempStoriesType.Author = story.User.Username
						tempStoriesType.Summarize = resp
						if rand.Float32() < 0.33 {
							var tempAsset Asset
							tempAsset.Type = "image"
							tempAsset.Src = media.URL
							medias = append(medias, tempAsset)
						}
						if profile.User.Friendship.FollowedBy {
							temp = append([]openai.StoriesType{tempStoriesType}, temp...)
						} else {
							temp = append(temp, tempStoriesType)
						}
					}

					break
				}
				if reachedLimit {
					break
				}
			}
			summarize, err := openai.SummarizeImagesToOne(temp)
			if err != nil {
				continue
			}
			if summarize != "Nothing interesting" {
				today := time.Now().Format("02.01.2006")

				todayExists := false
				for _, entry := range thisWeek {
					if entryContainsDate(entry, today) {
						todayExists = true
						break
					}
				}
				if !todayExists {
					thisWeek = append(thisWeek, summarize+" "+today)
					if len(thisWeek) > 7 {
						thisWeek = thisWeek[len(thisWeek)-7:]
					}
					stringified, err := json.Marshal(thisWeek)
					if err != nil {
						log.Println(err)
						stringified = []byte(data)
					}
					_ = redis.StoreInRedis(context.Background(), story.User.Username, string(stringified), 7*24*time.Hour)
				}
				var usersStories openai.StoriesType
				usersStories.Author = story.User.Username
				usersStories.Summarize = summarize
				storiesArray = append(storiesArray, usersStories)
			}
		}
	}
	data, err := GenerateVideoJson(medias)
	if err != nil {
		log.Println(err)
		return used, "", storiesArray
	}
	id, err := GenerateVideo(data)
	if err != nil {
		log.Println(err)
		return used, "", storiesArray
	}
	url, err := GetUrl(id)
	if err != nil {
		log.Println(err)
		return used, "", storiesArray
	}

	return used, url, storiesArray
}

func entryContainsDate(entry, date string) bool {
	return len(entry) > 10 && entry[len(entry)-10:] == date
}
