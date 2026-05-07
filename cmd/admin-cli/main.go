package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/asey0u/attendance-service/internal/admin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../../.env")

	var serverURL string
	flag.StringVar(&serverURL, "server", "", "attendance server URL")
	flag.StringVar(&serverURL, "s", "", "attendance server URL (shorthand)")
	flag.Parse()

	if serverURL == "" {
		serverURL = "http://" + os.Getenv("APP_HOST") + ":" + os.Getenv("APP_PORT")
	}

	cli := admin.NewCLI(serverURL)
	if err := cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
