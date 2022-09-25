package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var operation string

	flag.StringVar(&operation, "o", "", "Operation")

	flag.Parse()

	switch operation {
	case "register":
		hcpRegisterWallet()
	default:
		fmt.Println("Invalid Argument value! Operation argument value does not supported.")
		os.Exit(1)
	}
}
