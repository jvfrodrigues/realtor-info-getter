package c2s

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jvfrodrigues/realtor-info-getter/pkg/shared"
)

func GetC2S(args []string) {
	DEFAULT_FOLDER_PATH := "./c2s_clients"
	username := ""
	password := ""
	folderPath := DEFAULT_FOLDER_PATH
	if len(args) >= 2 {
		username = args[0]
		password = args[1]
        if len(args) > 2 {
    		folderPath = args[2]
        }
	} else {
		// authenticate
		username = shared.StringPrompt("username:")
		if username == "" {
			log.Panicln("User cant be nil")
		}
		password = shared.StringPrompt("password:")
		if password == "" {
			log.Panicln("Password cant be nil")
		}
		folderPath = shared.StringPrompt("save to folder (default ./c2s_clients):")
	}

	if folderPath == "" {
		folderPath = DEFAULT_FOLDER_PATH
	}

	if _, err := os.Stat(folderPath); errors.Is(err, fs.ErrNotExist) {
		os.MkdirAll(folderPath, os.ModePerm)
	}

	c := colly.NewCollector()
	start := time.Now()

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
				path := strings.Split(nextPageURL, "/")
				fmt.Println(path[3])
				fileName := fmt.Sprintf("%s/%s.html", folderPath, path[3])
				if _, err := os.Stat(fileName); errors.Is(err, fs.ErrNotExist) {
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
			fileName := fmt.Sprintf("%s/%s.html", folderPath, path[5])
			err = os.WriteFile(fileName, r.Body, 0o644)
			if err != nil {
				log.Fatal("ERROR WRITING TO FILE ", err)
			}
			fmt.Printf("HTML content saved to %s\n", fileName)

		}
	})

	c.Visit("https://api.contact2sale.com/webapp/leads/v2")
	elapsed := time.Since(start)
	log.Printf("took %s", elapsed)
}
