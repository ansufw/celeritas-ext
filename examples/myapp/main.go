package main

import (
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"
	"os"

	"github.com/ansufw/celeritas"
)

type application struct {
	App        *celeritas.Celeritas
	handlers   *handlers.Handlers
	models     *data.Model
	Middleware *middleware.Middleware
}

func main() {
	c := initApplication()
	err := c.App.ListenAndServe()
	if err != nil {
		c.App.ErrorLog.Println(err)
		os.Exit(1)
	}
}
