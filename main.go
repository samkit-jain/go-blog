// Package main is the entry point of the app
package main

import (
	"net/http"

	"github.com/samkit-jain/go-blog/api"
	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/website"
)

// App is the main super handler that routes URLs to their specific handlers
type App struct {
	WebsiteHandler *website.WebsiteHandler
	ApiHandler     *api.ApiHandler
}

// App's ServeHTTP routes URLs to the handlers for API and website
func (h *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, _ = helpers.ShiftPath(req.URL.Path)

	if head == "api" {
		_, req.URL.Path = helpers.ShiftPath(req.URL.Path)

		h.ApiHandler.ServeHTTP(res, req)
	} else {
		h.WebsiteHandler.ServeHTTP(res, req)
	}

	return
}

func main() {
	// initialise database connection
	config.InitDB()

	// initialise main handler
	app := &App{
		ApiHandler:     api.NewApiHandler(),
		WebsiteHandler: website.NewWebsiteHandler(),
	}

	// start listening
	http.ListenAndServe(":8080", app)
}
