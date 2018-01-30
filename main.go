package main

import (
	"net/http"

	"github.com/samkit-jain/go-blog/api"
	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/website"
)

type App struct {
	WebsiteHandler *website.WebsiteHandler
	ApiHandler     *api.ApiHandler
}

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
	config.InitDB()

	app := &App{
		ApiHandler:     new(api.ApiHandler),
		WebsiteHandler: new(website.WebsiteHandler),
	}

	http.ListenAndServe(":8080", app)
}
