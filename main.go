package main

import (
	"net/http"

	_ "github.com/lib/pq"

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
