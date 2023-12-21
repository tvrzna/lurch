package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/websocket"
)

func main() {
	c := NewContext(LoadConfig(os.Args))

	if c.conf.client {
		os.Exit(makeSocketClientAction(c.conf.action, c.conf.data))
	}

	go runWebServer(c)

	handleStop(c)
}

func runWebServer(c *Context) {
	mux := http.NewServeMux()

	mux.Handle("/ws/", websocket.Handler(NewWebSocketService(c).HandleWebSocket))
	mux.HandleFunc("/rest/", (&RestService{c: c}).HandleFunc)
	mux.HandleFunc("/", NewWebService(c).HandleFunc)
	c.webServer = &http.Server{Handler: mux, Addr: c.conf.getServerUri()}

	log.Print("-- lurch started on ", c.conf.getServerUri())
	if err := c.webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print("-- lurch start failed: ", err)
		c.interrupt <- true
	} else {
		log.Print("-- lurch finished")
	}
}

func handleStop(c *Context) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	select {
	case <-ch:
	case <-c.interrupt:
	}

	log.Print("-- stopping lurch")
	c.InterruptAll()
	if c.webServer != nil {
		c.webServer.Close()
	}
}
