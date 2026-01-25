package main

import (
	"log"
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"
	"os"

	"github.com/ansufw/celeritas"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// init celeritas
	cel := &celeritas.Celeritas{}
	if err := cel.New(path); err != nil {
		log.Fatal(err)
	}

	cel.AppName = "MyApp"

	myMiddleware := &middleware.Middleware{
		App: cel,
	}

	cel.InfoLog.Println("Debug is set to: ", cel.Debug)

	myHandlers := &handlers.Handlers{
		App:    cel,
		Models: data.New(cel.DB.Pool),
	}

	app := &application{
		App:        cel,
		handlers:   myHandlers,
		Middleware: myMiddleware,
	}

	app.App.Routes = app.routes()

	app.models = data.New(app.App.DB.Pool)

	app.Middleware.Models = app.models

	return app
}
