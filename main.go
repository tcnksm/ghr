package main

import (
	"log"
	"os"
)

func main() {
	cli := &CLI{outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(cli.Run(os.Args))
}

// Debug display values when DEBUG mode
// This is used only for developer
func Debug(v ...interface{}) {
	if os.Getenv("GHR_DEBUG") != "" {
		log.Println(v...)
	}
}
