package main

import (
	"os"
	"os/signal"
	"syscall"
	"xtimer/app"
)

func main() {

	webServer := app.GetWebServer()

	webServer.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	<-quit

}
