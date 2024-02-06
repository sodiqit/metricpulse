package main

import "flag"

type Flags struct {
	serverAddr     string
	reportInterval int
	pollInterval   int
}

var agentFlags Flags

func parseFlags() {
	flag.StringVar(&agentFlags.serverAddr, "a", "localhost:8080", "address and port server")
	flag.IntVar(&agentFlags.reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&agentFlags.pollInterval, "p", 2, "poll runtime interval in seconds")

	flag.Parse()
}
