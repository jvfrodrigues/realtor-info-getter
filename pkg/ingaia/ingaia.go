package ingaia

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jvfrodrigues/realtor-info-getter/pkg/shared"
)

var (
	BEARER_TOKEN string
	FOLDER_PATH  string
	COOKIES      []*http.Cookie
)

type IngaiaResponse struct {
	Hits  []IngaiaResponseItem `json:"hits"`
	Total int                  `json:"total"`
}

type Photos struct {
	Big string `json:"big"`
}

type IngaiaResponseItem struct {
	ID                    int      `json:"id"`
	LocationStreetNumber  int      `json:"location_street_number"`
	Usage                 string   `json:"usage"`
	RentPrice             float64  `json:"rent_price"`
	Photos                []Photos `json:"photos"`
	RentAveragePrice      float64  `json:"rent_average_price"`
	MunicipalPropertyTax  float64  `json:"municipal_property_tax"`
	UsageType             string   `json:"usage_type"`
	PropertyReference     string   `json:"property_reference"`
	LocationStreetAddress string   `json:"location_street_address"`
	Bathrooms             int      `json:"bathrooms"`
	Garages               int      `json:"garages"`
	BedroomBath           int      `json:"bedroom_bath"`
	SaleAveragePrice      float64  `json:"sale_average_price"`
	LocationAddOnAddress  string   `json:"location_add_on_address"`
	LocationNeighborhood  string   `json:"location_neighborhood"`
	Beds                  int      `json:"beds"`
	AreaUseful            float64  `json:"area_useful"`
	Enterprise            string   `json:"enterprise"`
	LocationState         string   `json:"location_state"`
	TotalGarages          int      `json:"total_garages"`
	AreaBuilt             float64  `json:"area_built"`
	LocationCity          string   `json:"location_city"`
	HasNegotiation        bool     `json:"has_negotiation"`
	CondoPrice            float64  `json:"condo_price"`
	AgencyName            string   `json:"agency_name"`
	SalePrice             float64  `json:"sale_price"`
	HasProposal           bool     `json:"has_proposal"`
	Area                  float64  `json:"area"`
	AreaLabel             string   `json:"area_label"`
}

func GetIngaia(args []string) {
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

	currentPage := 0
	total := 0
	for {
		pageList := getPage(currentPage)
		if total == 0 {
			total = pageList.Total
		}
		for _, hit := range pageList.Hits {
			dir := filepath.Join(FOLDER_PATH, hit.PropertyReference)
			if _, err := os.Stat(dir); !errors.Is(err, fs.ErrNotExist) {
				continue
			}
			fmt.Println("new property")
			getItemInfoAsync(hit, dir)
		}
		if currentPage > total/36 {
			break
		}
		currentPage++
	}
	elapsed := time.Since(start)
	log.Printf("got %d pages, took %s", currentPage, elapsed)
}

func getPage(page int) IngaiaResponse {
	url := fmt.Sprintf("https://listings.ingaia.com.br/listings?per_page=36&sort_by=register_date&page_num=%d&scope=Agency&status_id=1", page)

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

func getItemInfoAsync(item IngaiaResponseItem, dir string) bool {
	var wg sync.WaitGroup
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return false
	}

	os.MkdirAll(dir, os.ModePerm)

	wg.Add(1)
	go getInfoAsync(&wg, item, dir)
	for index, photo := range item.Photos {
		wg.Add(1)
		go downloadPhoto(&wg, dir, item.PropertyReference, index, photo.Big)
	}
	wg.Wait()
	return true
}

func downloadPhoto(wg *sync.WaitGroup, dir, propertyReference string, index int, url string) {
	defer wg.Done()

	response, err := http.Get(url)
	if err != nil {
		log.Printf("Error downloading photo: %v", err)
		return
	}
	defer response.Body.Close()

	fileName := fmt.Sprintf("%s/%s_%d.jpg", dir, propertyReference, index)
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Printf("Error copying file: %v", err)
		return
	}
	fmt.Printf("Downloaded: %s\n", fileName)
}

func getInfoAsync(wg *sync.WaitGroup, item IngaiaResponseItem, dir string) {
	defer wg.Done()
	url := fmt.Sprintf("https://imob.valuegaia.com.br/admin/json/imovel/ficha/ficha.ashx?id=%d", item.ID)

	req, _ := http.NewRequest("GET", url, nil)

	for _, cookie := range COOKIES {
		req.AddCookie(cookie)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://imob.valuegaia.com.br/admin/modules-react/")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Sec-Fetch-Dest", "iframe")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-User", "?1")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("ERROR ON READ ", err)
	}

	defer res.Body.Close()
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

	if len(body) <= 0 {
		log.Panicln("Not getting HTML")
	}

	fileName := fmt.Sprintf("%s/info.html", dir)
	err = os.WriteFile(fileName, body, 0o644)
	if err != nil {
		log.Fatal("ERROR WRITING TO FILE ", err)
	}
	fmt.Printf("HTML content saved to %s\n", fileName)
}
