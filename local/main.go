package main

import (
	"embed"
	"encoding/json"
	"fmt"

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
	_, err = arbor.GetArborData(&c)
	if err != nil {
		fmt.Println("error:", err)
	}
}
