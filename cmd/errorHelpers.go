package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

//Sends a generic 500 Internal Server Error to client.
//Writes error message and stack trace to error log.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace) //Outputs trace with a depth of 2 so that it starts at where the error actually is.

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

//Sends specific error code to client.  Generic and flexible error handling.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
