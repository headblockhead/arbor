package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/headblockhead/arbor"
	"github.com/headblockhead/waveshareCloud"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "6868"
	CONN_TYPE = "tcp"
)

//go:embed creds.json
var creds []byte

// This variable exists to make sure "embed" is used which is required for the above line.
var redundant embed.FS

func main() {
	var c arbor.Creds
	err := json.Unmarshal(creds, &c)
	if err != nil {
		fmt.Printf("error parsing creds json: %v", err)
	}
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Printf("Error listening for connections: %v", err)
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	defer fmt.Println("Listener closed.")
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Error accepting a connection: %v", err)
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn, &c)
	}
}

func handleRequest(conn net.Conn, c *arbor.Creds) {
	fmt.Println("New connection from:", conn.RemoteAddr())
	// Setting up the connection to the display.
	lc := waveshareCloud.NewLoggingConn(conn, false)

	// Creating the display. If a password is required to unlock the display, here is where you would enter it.
	// This automatically unlocks the display when created.
	// In this case, the display is not locked, so the password is not required.
	display := waveshareCloud.NewDisplay(lc, c.DevicePassword)

	id, err := display.GetID()
	if id != c.DeviceID {
		fmt.Printf("Device ID Incorrect, stopping connection")
	}

	fmt.Println("Finding browser...")
	path, _ := launcher.LookPath()
	fmt.Println("Browser found:", path)
	u := launcher.New().Bin(path).Set("no-sandbox").MustLaunch()
	b := rod.New().ControlURL(u).MustConnect()

	data, err := arbor.GetArborData(c, b, true)
	img, err := arbor.GetArborImage(&data)
	if err != nil {
		fmt.Printf("Error getting arbor image: %v", err)
	}
	err = display.SendImage(img)
	if err != nil {
		fmt.Printf("Error sending testpattern image: %v\n", err)
	}

	// Shutdown the display.
	err = display.Shutdown()
	if err != nil {
		fmt.Printf("Error shutting down: %v\n", err)
	}
	// Close the connection.
	display.Disconnect()
}
