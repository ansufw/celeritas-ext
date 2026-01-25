package middleware

import (
	"myapp/data"

	"github.com/ansufw/celeritas"
)

type Middleware struct {
	App    *celeritas.Celeritas
	Models *data.Model
}
