package main

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/headblockhead/arbor"
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
	fmt.Println("Finding browser...")
	path, _ := launcher.LookPath()
	fmt.Println("Browser found:", path)
	u := launcher.New().Bin(path).Set("no-sandbox").MustLaunch()
	b := rod.New().ControlURL(u).MustConnect()
	_, err = arbor.GetArborData(&c, b, true)
	if err != nil {
		fmt.Println("error:", err)
	}
}
