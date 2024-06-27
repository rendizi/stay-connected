package main

import (
	"fmt"
	"github.com/MadAppGang/httplog"
	"net/http"
	"os"
	daily "stay-connected/internal/daily-fetcher"
	"stay-connected/internal/handler"
	"stay-connected/internal/server"
	"stay-connected/internal/services/db"
)

func main() {
	db.InitSupabase()

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/register", httplog.Logger(http.HandlerFunc(handler.Register)))
	mux.Handle("POST /api/v1/login", httplog.Logger(http.HandlerFunc(handler.Login)))
	mux.Handle("PUT /api/v1/credentials", httplog.Logger(http.HandlerFunc(handler.UpdateCredentials)))
	mux.Handle("DELETE /api/v1/credentials", httplog.Logger(http.HandlerFunc(handler.DeleteCredentials)))
	mux.Handle("/health", httplog.Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.Ok(map[string]interface{}{"message": "alright"}, w)
	})))

	//users, err := db.Insert(os.Getenv("TEST_INSTAGRAM_USERNAME"),os.Getenv("TEST_INSTAGRAM_PASSWORD"), "baglanov.a0930@gmail.com")
	//if err != nil{
	//	log.Fatal(err)
	//}
	//log.Println(users)

	go daily.Do()
	//err = mailer.Send("alikhan2008ba@gmail.com", "something")
	//log.Fatal(err)

	//s := gocron.NewScheduler()
	//s.Every(1).Days().Do(daily.Do)
	//<-s.Start()

	//fullusers, err := db.Get()
	//if err != nil{
	//	log.Fatal(err)
	//}
	//log.Println(fullusers)

	//return

	//l := stories.SummarizeInstagramStories(inst)
	//log.Println(l[0])

	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	fmt.Println("server is listening")

	err := http.ListenAndServe(":8080", corsHandler(mux))
	fmt.Println(err)
	if err != nil {
		if err == http.ErrServerClosed {
			fmt.Println("server closed")
		} else {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}
}
