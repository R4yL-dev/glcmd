package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/R4yL-dev/glcmd/internal/app"
)

func main() {
	client, err := makeHttpClient()
	if err != nil {
		log.Fatal(err)
	}

	email := os.Getenv("GL_EMAIL")
	if email == "" {
		log.Fatal("GL_EMAIL environment variable is not set")
	}

	password := os.Getenv("GL_PASSWORD")
	if password == "" {
		log.Fatal("GL_PASSWORD environment variable is not set")
	}

	glcmd, err := app.NewApp(email, password, client)
	if err != nil {
		log.Fatal(err)
	}

	gm, err := glcmd.GetMeasurement()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gm.ToString())
}

func makeHttpClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
	return client, nil
}
