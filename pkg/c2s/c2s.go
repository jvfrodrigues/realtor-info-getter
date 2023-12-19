package c2s

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jvfrodrigues/realtor-info-getter/pkg/shared"
)

func GetC2S() {
	FOLDER_PATH := "./c2s_clients"
	folderPath := shared.StringPrompt("save to folder (default ./c2s_clients):")
	if folderPath != "" {
		FOLDER_PATH = folderPath
	}
	if _, err := os.Stat(folderPath); errors.Is(err, fs.ErrNotExist) {
		os.MkdirAll(folderPath, os.ModePerm)
	}
	c := colly.NewCollector()

	// authenticate
	username := shared.StringPrompt("username:")
	if username == "" {
		log.Panicln("Bearer cant be nil")
	}
	password := shared.StringPrompt("password:")
	if password == "" {
		log.Panicln("Bearer cant be nil")
	}

	err := c.Post("https://api.contact2sale.com/webapp/login/login", map[string]string{"username": username, "password": password, "redirect_to": ""})
	if err != nil {
		log.Fatal(err)
	}

	// attach callbacks after login
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})

	c.OnHTML("div#tab-all a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Attr("href")
		if strings.Contains(nextPageURL, "hash") {
			fmt.Println(nextPageURL)
			if nextPageURL != "" {
				absURL := e.Request.AbsoluteURL(nextPageURL)
				if absURL == "" {
					log.Println("Error resolving URL:", err)
					return
				}
				err = c.Visit(absURL)
				if err != nil {
					log.Println("Error visiting next page:", err)
				}
			}
		}
	})

	c.OnHTML("div#tab-all li.next a[rel='next']", func(e *colly.HTMLElement) {
		nextPageURL := e.Attr("href")
		fmt.Println(nextPageURL)
		if nextPageURL != "" {
			absURL := e.Request.AbsoluteURL(nextPageURL)
			if absURL == "" {
				log.Println("Error resolving URL:", err)
				return
			}
			err = c.Visit(absURL)
			if err != nil {
				log.Println("Error visiting next page:", err)
			}
		}
	})

	c.OnResponse(func(r *colly.Response) {
		url := r.Request.URL.String()
		if strings.Contains(url, "hash") {
			path := strings.Split(url, "/")
			fileName := fmt.Sprintf("%s/%s.html", FOLDER_PATH, path[5])
			err = os.WriteFile(fileName, r.Body, 0644)
			if err != nil {
				log.Fatal("ERROR WRITING TO FILE ", err)
			}
			fmt.Printf("HTML content saved to %s\n", fileName)
		}
	})

	c.Visit("https://api.contact2sale.com/webapp/leads/v2")
}
