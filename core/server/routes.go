package server

import (
	"sumeru/core/server/web"
)

func registerAppRoutes() {
	web.RegisterAppRoutes(nil)
}

func registerSetupRoutes() {
	web.RegisterSetupRoutes(nil)
}
