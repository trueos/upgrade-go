package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"time"

	"github.com/gorilla/websocket"
)

func rotatelog() {
	if _, err := os.Stat(logfile) ; os.IsNotExist(err) {
		return
	}
	cmd := exec.Command("mv", logfile, logfile + ".previous")
	cmd.Run()
}

func logtofile(info string) {
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(info + "\n")); err != nil {
		log.Fatal(err)
	}
	if err:= f.Close(); err != nil {
		log.Fatal(err)
	}
}

// Start the websocket server
func startws() {
        log.SetFlags(0)
        http.HandleFunc("/ws", readws)
        log.Fatal(http.ListenAndServe(*addr, nil))
}

// Start our client connection to the WS server
var (
        c   *websocket.Conn
)
func connectws() {
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
//	log.Printf("connecting to %s", u.String())

	err := errors.New("")
	var connected bool = false
	for attempt := 0; attempt < 10; attempt++ {
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			connected = true
			break
		}
		time.Sleep(5 * time.Millisecond);
	}
	if (!connected) {
		log.Fatal("Failed connecting to websocket server", err)
	}
}

// Called when we want to signal that its time to close the WS connection
func closews() {
	log.Println("Closing WS connection")
	defer c.Close()

	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}

func checkuid() {
	user, _ := user.Current()
	if ( user.Uid != "0" ) {
		fmt.Println("ERROR: Must be run as root")
		os.Exit(1)
	}
}

func main() {

	checkuid()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Capture any sigint
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		os.Exit(1)
	}()

	// Load the local config file if it exists
	loadconfig()

	if ( listtrainflag ) {
		go startws()
		connectws()
		listtrains()
		closews()
		os.Exit(0)
	}

	if ( changetrainflag != "" ) {
		go startws()
		connectws()
		listtrains()
		closews()
		os.Exit(0)
	}

	if ( checkflag ) {
		go startws()
		connectws()
		startcheck()
		closews()
		os.Exit(0)
	}

	if ( updateflag ) {
		go startws()
		connectws()
		startupdate()
		closews()
		os.Exit(0)
	}

	if ( stage2flag ) {
		startstage2()
		os.Exit(0)
	}

	if ( websocketflag ) {
		startws()
		os.Exit(0)
	}
}
