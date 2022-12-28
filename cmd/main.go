package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"flag"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	pgxpool "github.com/jackc/pgx/v4/pgxpool"
	godotenv "github.com/joho/godotenv"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
	"github.com/rainbowriverrr/F3Ytwitter/internal/validation"
)

type followRequest struct {
	User     *models.SimpleRequest
	upstream chan []*models.Follow
}
type connectionsRequest struct {
	ID      int64
	follows []*models.Follow `json:"follows"`
	//expects: "followings" or "followers"
	//"followings" means the slice of follows is the slice of followings
	//"followers" means the slice of follows is the slice of followers
	users string `json:'follower"`
}
type simplifiedSchool struct {
	Name          string `json:"name"`
	TopRated      bool   `json:"top_rated"`
	Public        bool   `json:"public"`
	City          string `json:"city"`
	State         string `json:"state_province"`
	Country       string `json:"country"`
	TwitterHandle string `json:"twitter_handle"`
}

// User structure used to pass data between goroutines.  This is used by the ProfileWorker, FollowerWorker, FollowingWorker, and TweetsWorker goroutines.
type simplifiedUser struct {
	ID                  int64             `json:"id"`
	Username            string            `json:"username"`
	IsSchool            bool              `json:"is_school"`
	IsParticipant       bool              `json:"isParticipant"`
	ParticipantSchoolID int               `json:"participantSchool"`
	ParticipantCohort   int               `json:"participantCohort"`
	SchoolInfo          *simplifiedSchool `json:"schoolInfo"`
	StartDate           time.Time         `json:"startDate"`
	ScrapeConnections   bool              `json:"scrape_connections"`
	ScrapeContent       bool              `json:"scrape_content"`
	BackupID            int64             `json:"backup_id"`
}

// Application dependencies to be injected
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	connection    *pgxpool.Pool
	scraper       twitterscraper.Scraper
	templateCache map[string]*template.Template
	//currently, automatic form decoding not implemented
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	debug          bool
	bearerToken    string
	bearerToken2   string
	apiKey         string
	secretKey      string
	followerClient http.Client
	//expects simplifiedUser struct
	profileChan     chan *simplifiedUser
	followChan      chan *models.SimpleRequest
	followerChan    chan *models.SimpleRequest
	tweetsChan      chan *simplifiedUser
	connectionsChan chan connectionsRequest
	followQueue     chan *followRequest
	followerQueue   chan *followRequest
	//statuses of channels
	profileStatus     string
	followStatus      string
	followingStatus   string
	connectionsStatus string
	tweetsStatus      string
	//the limit of the number of followers to scrape.  If the number of followers is greater than this, the followers will not be scraped.
	followLimit int
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

	//Loading Bearer Tokens and API keys
	infoLog.Println("Loading Bearer Tokens...")
	bearerToken := os.Getenv("BEARER_TOKEN")
	bearerToken2 := os.Getenv("BEARER_TOKEN2")
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	//Connects to the database using .env variables
	infoLog.Println("Connecting to database...")
	dburl := "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASS") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME")
	conn, err := pgxpool.Connect(context.Background(), dburl)
	if err != nil {
		errLog.Fatal(err)
	}
	defer conn.Close()
	infoLog.Println("Connected to database")

	//Loads web address and sets up server structs for dependency injection
	defaultAddr := os.Getenv("WEB_ADDR")
	addr := flag.String("addr", defaultAddr, "HTTP network address")

	//Initializes template cache
	infoLog.Println("Initializing template cache...")
	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	//Initializes form decoder
	infoLog.Println("Initializing form decoder...")
	formDecoder := form.NewDecoder()

	//Initializes session manager
	infoLog.Println("Initializing session manager...")
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(conn)
	sessionManager.Lifetime = 12 * time.Hour

	//Initializes http clients
	infoLog.Println("Initializing http clients...")
	followerClient := http.Client{}

	//Initializing channels
	infoLog.Println("Initializing channels...")
	profileChan := make(chan *simplifiedUser, 100)
	followChan := make(chan *models.SimpleRequest, 3000)
	followerChan := make(chan *models.SimpleRequest, 3000)
	tweetsChan := make(chan *simplifiedUser, 100)
	connectionsChan := make(chan connectionsRequest, 100)
	followQueue := make(chan *followRequest, 1000)
	followerQueue := make(chan *followRequest, 1000)

	defer close(profileChan)
	defer close(followChan)
	defer close(followerChan)
	defer close(tweetsChan)
	defer close(connectionsChan)
	defer close(followQueue)
	defer close(followerQueue)

	profileStatus := "idle"
	followStatus := "idle"
	followingStatus := "idle"
	tweetsStatus := "idle"
	connectionsStatus := "idle"

	app := &application{
		errorLog:          errLog,
		infoLog:           infoLog,
		connection:        conn,
		scraper:           *twitterscraper.New(),
		debug:             false,
		templateCache:     templateCache,
		formDecoder:       formDecoder,
		sessionManager:    sessionManager,
		bearerToken:       bearerToken,
		bearerToken2:      bearerToken2,
		apiKey:            apiKey,
		secretKey:         secretKey,
		followerClient:    followerClient,
		profileChan:       profileChan,
		followChan:        followChan,
		followerChan:      followerChan,
		tweetsChan:        tweetsChan,
		connectionsChan:   connectionsChan,
		followQueue:       followQueue,
		followerQueue:     followerQueue,
		profileStatus:     profileStatus,
		followStatus:      followStatus,
		followingStatus:   followingStatus,
		tweetsStatus:      tweetsStatus,
		connectionsStatus: connectionsStatus,
		followLimit:       1000,
	}

	//Initializes concurrent workers
	infoLog.Println("Initializing concurrent workers...")
	go app.ProfileWorker()
	go app.FollowWorker()
	go app.FollowerWorker()
	go app.TweetsWorker()
	go app.ConnectionsWorker()
	go app.FollowerQueue()
	go app.FollowingQueue()

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errLog,
		Handler:  app.routes(),
	}

	//Basic CLI controls
	fmt.Printf("\n~~Hello, welcome to the F3Y Twitter Scraper!~~\n")
	fmt.Printf("~~This program will scrape twitter profiles and store them in a database.~~\n")
	fmt.Printf("~~To select a command, ender it's number below.~~\n")
	fmt.Printf("~~You can use the following commands to interact with the program:~~\n")

	reader := bufio.NewReader(os.Stdin)
	choosing := true
	for choosing {
		fmt.Printf("\n 1) Initialize/Reset Tables (Warning, all data will be lost)")
		fmt.Printf("\n 2) Start Web Server")
		fmt.Printf("\n 3) Scrape User and add to Database")
		fmt.Printf("\n 4) List all users in Database")
		fmt.Printf("\n 5) Add Admin User")
		fmt.Printf("\n 6) Quit")
		fmt.Printf("\n")

		char, _, err := reader.ReadRune()
		if err != nil {
			errLog.Println(err)
			continue
		}
		reader.Reset(os.Stdin)
		switch char {
		case '1':
			fmt.Printf("\n~~Initializing Tables~~\n")
			err := app.resetTables()
			if err != nil {
				errLog.Println(err)
			}
			fmt.Printf("\n~~Tables Initialized~~\n")
		case '2':
			fmt.Printf("\n~~Starting Web Server~~\n")
			choosing = false
		case '3':
			app.scrapeCLI(reader)
		case '4':
			fmt.Printf("\n~~Listing all users in Database~~\n")
			users, err := models.GetAllUsernames(app.connection)
			if err != nil {
				errLog.Println(err)
				continue
			}
			for _, user := range users {
				fmt.Printf("%s\n", user)
			}
		case '5':
			app.addAdmin(reader)
		case '6':
			fmt.Printf("\n~~Quitting~~\n")
			os.Exit(0)
		}
		reader.Reset(os.Stdin)

	}

	app.infoLog.Printf("Starting server on %s...", *addr)
	err = srv.ListenAndServe()
	errLog.Fatal(err)

}

func (app *application) scrapeCLI(r *bufio.Reader) {
	fmt.Printf("\n~~Please enter a twitter username to scrape~~\n")
	username, _ := r.ReadString('\n')
	username = username[:len(username)-1]
	user, err := app.scrapeUser(username)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	err = app.addOrUpdateUser(user)
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) addAdmin(r *bufio.Reader) {
	fmt.Printf("\n~~Please enter an email to add as an admin~~\n")
	username, _ := r.ReadString('\n')
	username = username[:len(username)-1]
	//checks if username is valid email address
	if !validation.Matches(username, validation.EmailEXP) {
		fmt.Printf("\n~~%s is not a valid email address.~~\n", username)
		return
	}
	r.Reset(os.Stdin)
	fmt.Printf("\n~~Please enter a password for the admin account~~\n")
	password, _ := r.ReadString('\n')
	password = password[:len(password)-1]
	//checks if password is longer than 4 characters
	if len(password) < 4 {
		fmt.Printf("\n~~Password must be longer than 4 characters.~~\n")
		return
	}
	//adds admin
	err := models.InsertAdmin(app.connection, &models.Admin{Email: username, Password: []byte(password)})
	if err != nil {
		app.errorLog.Println(err)
	}
}
