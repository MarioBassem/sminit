package main

import (
	"log"
	"os"

	"github.com/mariobassem/sminit-go/swatcher"
)

func main() {
	err := swatcher.Swatch()
	if err != nil {
		log.Print(err)
	}
	cleanUp()
}

// cleanUp should delete /run/sminit
func cleanUp() {
	err := os.RemoveAll(swatcher.SminitRunDir)
	if err != nil {
		swatcher.SminitLogFail.Print(err)
	}
}
