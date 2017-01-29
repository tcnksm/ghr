package main

import "os"

// envDebug is used for changing verbose outoput
var envDebug = "DEBUG"

func main() {
	cli := &CLI{outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(cli.Run(os.Args))
}
