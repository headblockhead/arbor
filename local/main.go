package main

import (
	"fmt"

	"github.com/headblockhead/arbor"
)

func main() {
	_, err := arbor.GetArborData()
	if err != nil {
		fmt.Println("error:", err)
	}
}
