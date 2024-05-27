package ingaia

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jvfrodrigues/realtor-info-getter/pkg/shared"
)

const ACTIVE_STATUS string = "Ativo"

func UpdateIngaia(args []string) {
	username := ""
	password := ""
	if len(args) >= 2 {
		username = args[0]
		password = args[1]
		if len(args) > 2 {
			FOLDER_PATH = args[2]
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
		FOLDER_PATH = shared.StringPrompt("FOLDER PATH (default ./ingaia_listings):")
	}

	if FOLDER_PATH == "" {
		FOLDER_PATH = "./ingaia_listings"
	}

	start := time.Now()
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Status Code:", r.StatusCode)
		setCookieHeaders := r.Headers.Values("set-cookie")
		fmt.Println("Set-Cookie Headers:", setCookieHeaders)
		COOKIES = c.Cookies(r.Request.URL.String())
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Headers.Set("Origin", "https://signin.valuegaia.com.br")
		r.Headers.Set("Referer", "https://signin.valuegaia.com.br")
	})

	err := c.Post("https://imob.valuegaia.com.br/login-valida.aspx", map[string]string{"txLogin": username, "txSenha": password, "txWiki=": "", "txGaiaInc=": "", "txPage=": "", "txValue=": "", "txEid=": "", "txGaiaAdsAccessToken=": "", "txGaiaAdsRefreshToken=": "", "key": "14df1ee61b55cda2b61b307ab9cc475c"})
	if err != nil {
		log.Fatal(err)
	}

	dir, err := os.ReadDir(FOLDER_PATH)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range dir {
		pageList := findListing(entry.Name())
		if pageList.Total == 0 {
			continue
		}
		for _, hit := range pageList.Hits {
			if hit.PropertyReference != entry.Name() {
				continue
			}
			fmt.Println(entry.Name(), hit.Status != ACTIVE_STATUS)
			if hit.Status != ACTIVE_STATUS && entry.IsDir() {
				os.Rename(FOLDER_PATH+"/"+entry.Name(), FOLDER_PATH+"/_"+entry.Name())
			}
		}
	}
	elapsed := time.Since(start)
	log.Printf("updated, took %s", elapsed)
}

func findListing(id string) IngaiaResponse {
	url := fmt.Sprintf("https://listings.ingaia.com.br/listings?per_page=36&page_num=0&scope=Agency&property_reference=%s", id)

	req, _ := http.NewRequest("GET", url, nil)
	var token string

	for _, cookie := range COOKIES {
		if cookie.Name == "accounts_token" {
			token = "Bearer " + cookie.Value
		}
	}

	req.Header.Add("cookie", "__goc_session__=olysoeyaoqswtxpebwxktooxxtvjfrqx")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referer", "https://imob.valuegaia.com.br/")
	req.Header.Add("Origin", "https://imob.valuegaia.com.br")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Authorization", token)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("TE", "trailers")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("ERROR ON DO ", err)
	}

	defer res.Body.Close()

	fmt.Println(res.Status)
	var body []byte
	if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
		reader, err := gzip.NewReader(res.Body)
		if err != nil {
			log.Fatal("ERROR ON GZIP NEW READER ", err)
		}
		defer reader.Close()

		// Read the gzipped content
		body, err = io.ReadAll(reader)
		if err != nil {
			log.Fatal("ERROR ON READ GZIP ", err)
		}
	} else {
		// If not gzip-encoded, read the response directly
		body, err = io.ReadAll(res.Body)
		if err != nil {
			log.Fatal("ERROR ON READ ", err)
		}
	}

	var data *IngaiaResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal("ERROR ON UNMARSHAL ", err)
	}
	return *data
}
