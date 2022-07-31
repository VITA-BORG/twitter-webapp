# F3Ytwitter
This is a basic web app which interacts with the F3Y database

## Requirements
This application has a few requirements.  This project uses Go 1.18.
This application also requires 
- Go 1.18
- .env File Requirements
  - DB_USER
  - DB_PASS
  - DB_NAME
  - DB_HOST
  - DB_PORT
  - WEB_ADDR
  - BEARER_TOKEN
  - BEARER_TOKEN
  - SECRET_KEY
  - API_KEY
The .env file provides a list of environment variables that you can use to change how the program connects to the database, what address the web server starts on, and important secret tokens that allows the scraper to obtain data from the twitter api.  To set up an environment file, create a file named .env in the root directory of the project.  The following code block is an example of the simple format that should be followed to create this file:
```
DB_USER=username_here
DB_PASS=pass_here
DB_NAME=db_name_here
DB_HOST=location_of_db
DB_PORT=1234
WEB_ADDR=:1234
```

### Project Structure

This project requires the following strucutre:


## Routes
There are a few routes currently implemented in the web app.

| Route                 | Description                                                                                                                         |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| /                     | The home/dashboard of the web application.  There is a overview of the system on this page                                          |
| /schools              | This page shows the schools that are already added in the system.  This page also contains the form to add schools into the system. |
| /users                | This provides an overview of the users currently added in the system                                                                |
| /users/view/:username | Every user in the system will have their own page that allows you to view and edit their information in the database                |
| /users/add            | The form to add participant users into the system                                                                                   |

## Running

Before you run this project, there are a few things you have to set up.

1. The .env file as outlined above
2. A postgresql database with its information placed in the .env file
   
If these are set up, you can run the project with:
```
go run ./cmd
```
You will see information about the status of the project, notifying you of any errors, such as an incorrectly formatted .env file.  If everything is in order, you will be greeted by a command line interface with a few options:

1. Initialize/Reset Tables
2. Start Web Server
3. Scrape User and add to Database
4. List all users in Database
5. Random testing option
6. Quit

If it is your first time running this application, you must first select option 1.  This will initialize the database to the proper structure.  After you do this, you will be able to select option 2 to start the web server.

Option 3 will allow you to add a user to the scrape, however currently this does not support adding schools or participants.

After yhou start the Web Server, you will need to add a school to the database in order to add participants connected to these schools.  To do this, navigate to the address that you provided in the .env file, and navigate to the schools page.  Here you will be able to add a school into the system.  After you do this, you navigate to the "Users" page and you will be able to start adding participants into the scrape.