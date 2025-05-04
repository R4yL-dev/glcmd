package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/R4yL-dev/glcmd/internal/app"
)

func main() {
	client, err := makeHttpClient()
	if err != nil {
		log.Fatal(err)
	}

	email := os.Getenv("GL_EMAIL")
	password := os.Getenv("GL_PASSWORD")

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
		Jar: jar,
	}
	return client, nil
}
