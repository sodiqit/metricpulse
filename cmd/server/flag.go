package main

import "flag"

var runAddr string

func parseFlags() {
	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")

	flag.Parse()
}
