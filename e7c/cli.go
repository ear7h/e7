package main

import (
	"os"
	"strconv"
	"fmt"
	"github.com/ear7h/e7/client"
)

func printHelp() {
	fmt.Println(`usage: e7 <command> <args>
commands:
	get - get an available port
	register <name> <port> - register the name and port`)
}

func main()  {
	// /path/to/ex <subcommand>
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "get":
		client.Get()
	case "register":
		// /path/to/ex register <name> <port>
		if len(os.Args) != 4 {
			printHelp()
			os.Exit(1)
		}
		name, port := os.Args[2], os.Args[3]
		ui, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = client.Register(name, int(ui))
		if err != nil {
			fmt.Println("Error registering")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(name, " successfully registered to ", port)
		return
	default:
		printHelp()
		os.Exit(1)
	}
}
