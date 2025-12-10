package main

import (
	"7dtd-monitor/internal/telnet"
	"7dtd-monitor/internal/ui"
	"flag"
	"fmt"
	"os"
)

func main() {
	host := flag.String("host", "localhost", "Server Host/IP")
	port := flag.String("port", "8081", "Telnet Port")
	password := flag.String("password", "", "Telnet Password")
	flag.Parse()

	if *password == "" {
		// In a real app we might prompt or error, but for mock default is empty ok
		// fmt.Println("Warning: No password provided")
	}

	client := telnet.NewClient(*host, *port, *password)
	app := ui.NewApp(client)

	if err := app.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
