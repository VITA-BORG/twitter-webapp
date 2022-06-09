package main

import (
	"context"
	"log"
	"os"

	//"net/http"
	//"flag"

	pgx "github.com/jackc/pgx/v4"
	godotenv "github.com/joho/godotenv"
	twitterscraper "github.com/n0madic/twitter-scraper"
)

type application struct {
	errorLog   *log.Logger
	infoLog    *log.Logger
	connection *pgx.Conn
	scraper    twitterscraper.Scraper
}

func main() {
	//Sets up Logs
	//TODO: allow change via command line flags
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
	//defaultAddr := os.Getenv("WEB_ADDR")
	//addr := flag.String("addr", defaultAddr, "HTTP network address")

	app := &application{
		errorLog:   errLog,
		infoLog:    infoLog,
		connection: conn,
		scraper:    *twitterscraper.New(),
	}

	currUser := scrapeUser(app, "NyameDev")
	infoLog.Println(currUser)

	// srv := &http.Server{
	// 	Addr:     *addr,
	// 	ErrorLog: errLog,
	// }
}
