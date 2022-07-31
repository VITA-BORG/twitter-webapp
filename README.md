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
