package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"flag"

	pgx "github.com/jackc/pgx/v4"
	godotenv "github.com/joho/godotenv"
	twitterscraper "github.com/n0madic/twitter-scraper"
	_ "github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

type application struct {
	errorLog   *log.Logger
	infoLog    *log.Logger
	connection *pgx.Conn
	scraper    twitterscraper.Scraper
	debug      bool
}

func main() {
	//Sets up Logs
	//TODO: allow changing logs via command line flags
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("Starting...")
	infoLog.Println("Loading env...")
	err := godotenv.Load()
	if err != nil {
		errLog.Println("Error loading .env file")
	}
	infoLog.Println("env loaded")

	//Connects to the database using .env variables
	infoLog.Println("Connecting to database...")
	dburl := "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASS") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME")
	conn, err := pgx.Connect(context.Background(), dburl)
	if err != nil {
		errLog.Fatal(err)
	}
	defer conn.Close(context.Background())
	infoLog.Println("Connected to database")

	//Loads web address and sets up server structs for dependency injection
	defaultAddr := os.Getenv("WEB_ADDR")
	addr := flag.String("addr", defaultAddr, "HTTP network address")

	app := &application{
		errorLog:   errLog,
		infoLog:    infoLog,
		connection: conn,
		scraper:    *twitterscraper.New(),
		debug:      true,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errLog,
		Handler:  app.routes(),
	}

	if app.debug {
		currUser, err := app.scrapeUser("nyamedev")
		if err != nil {
			errLog.Printf("Could not scrape user: %s, Error: %s", "nyamedev", err)
		}
		infoLog.Println(currUser.Gender)
		infoLog.Println(currUser.IsPerson)

		tweets := app.scrapeTweets(currUser.Handle, time.Date(2022, 4, 12, 0, 0, 0, 0, time.Local))
		mentionedUsers, mentions := app.scrapeMentions(tweets, true)
		replies := getReplies(tweets)

		for _, tweet := range tweets {
			app.infoLog.Printf("%s\n", tweet.Text)
		}
		app.infoLog.Printf("%d tweets scraped", len(tweets))
		app.infoLog.Printf("%d users mentioned", len(mentionedUsers))
		app.infoLog.Printf("%+v\n", mentionedUsers)
		app.infoLog.Printf("%d tweets with mentions", len(mentions))
		app.infoLog.Printf("%+v\n", mentions)
		app.infoLog.Printf("%d replies", len(replies))
		app.infoLog.Printf("%+v\n", getBioTags(currUser.Bio))
	}

	app.infoLog.Printf("Starting server on %s...", *addr)
	err = srv.ListenAndServe()
	errLog.Fatal(err)

}
