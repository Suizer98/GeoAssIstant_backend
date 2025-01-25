package main

import (
	"geoai-app/app"
)

func main() {
	var a app.App
	a.CreateConnection()
	a.Migrate()
	a.CreateRoutes()
	a.Run()
}
