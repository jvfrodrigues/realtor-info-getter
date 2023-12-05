package main

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

var TOKEN = ""

func main() {
	saveFolder := os.Getenv("FOLDER_PATH_AUTO")
	if saveFolder == "" {
		log.Panicln("Env var for path to folder not defined")
	}
	start := time.Now()
	args := os.Args
	if len(args) < 2 {
		log.Fatal("missing token")
	}
	TOKEN = args[1]
	currentPage := 0
	total := 0
	for {
		pageList := getPage(currentPage)
		if total == 0 {
			total = pageList.Total
		}
		for _, hit := range pageList.Hits {
			dir := filepath.Join(os.Getenv("FOLDER_PATH_AUTO"), hit.PropertyReference)
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
	req.Header.Add("Authorization", "Bearer "+TOKEN)
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

	req.Header.Add("cookie", "GaiaAuthenticationValues=MjIvMTEvMjAyMyAxNToxNToyMA%3D%3D")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://imob.valuegaia.com.br/admin/modules-react/")
	req.Header.Add("Cookie", "showNewMenu=false; SignInNoPlans=11/22/2023 14:05:41; _BEAMER_USER_ID_xcTSKWBK36426=875111a9-a7e7-4547-a241-e94ece7b66e3; _BEAMER_FIRST_VISIT_xcTSKWBK36426=2023-05-14T20:50:07.675Z; _BEAMER_LAST_PUSH_PROMPT_INTERACTION_xcTSKWBK36426=1684097411562; _BEAMER_BOOSTED_ANNOUNCEMENT_DATE_xcTSKWBK36426=2023-11-12T15:20:10.961Z; _BEAMER_DATE_xcTSKWBK36426=2023-10-24T19:27:47.193Z; _BEAMER_LAST_UPDATE_xcTSKWBK36426=1698170367584; __goc_session__=kiozfjdnqbffmarfuexgetsiyldjapgm; ASP.NET_SessionId=nb5sol5xm0vvlvydxdbxbpf2; B=FF4E05D034CABB212673E54C4AA19AB58EC5EA2B; GaiaAuthenticationLoginTime=11/22/2023 14:05:41; teste_https=value_teste_https; enterprises_url=https://enterprises.ingaia.com.br; signalr_url=https://chat.ingaia.com.br; chat_api_url=https://api-chat.ingaia.com.br; contacts_api_url=https://contacts-chat.ingaia.com.br; gaiacore_url=https://imob.gaiacore.com.br/; matomo=3; environment=Production; listings_url=https://listings.ingaia.com.br; accounts_url=https://accounts.ingaia.com.br; goals_url=https://imob-goals.ingaia.com.br; leads_url=https://leads.ingaia.com.br; clients_radar_url=https://radar.ingaia.com.br; customers_url=https://customers.ingaia.com.br; customer_search_url=https://customer-search.ingaia.com.br; kernel_url=https://kernel.ingaia.com.br; locations_url=https://locations.ingaia.com.br; integration_rd_url=https://integration-rdstation.ingaia.com.br/; rd_url=https://api.rd.services/auth/dialog?client_id=fe274f90-9def-4a76-a7b3-3edeafdc6e11; pipe_leads_url=https://pipe-leads.ingaia.com.br/; notifications_url=https://api-notifications.ingaia.com.br; notifications_socket_url=wss://notifications.ingaiaimobApiNotificationscom.br; page_dashboard_loan_url=https://credito.valuegaia.com.br; cms_authorize_page=https://sites-cms.kenlo.io/authorize; mf_onboarding_dashboard_loan_url=https://mf-broker-onboardings-z6rnix554q-uc.a.run.app; mf_kenlo_expert_presentation_url=https://mf-kenlo-expert-presentation-z6rnix554q-uc.a.run.app; config_lead_rules_mf_url=https://config-lead-rules.imob.stg.kenlo.io; home_equity_bff=https://us-central1-kenlo-kash.cloudfunctions.net/ms-bff-opportunities-prod-main; gaiacore_4=052902a8ea5976efa1bad600b7ec1b89; accounts_token="+TOKEN+"; GaiaAuthentication=82DFDD9273AFCCC9D2F18A928A446CB94F59F1D86EED38F7EE1C478BA5E52B8BCFAE0483CC18E16655ED6DCE09F26428E6ADEF090A0484F7FDE0A0E650730F4C167B4E2A6EA69EA8ACB71C9B2249156D3845118AA69AA439F00FF97A9CDAE57C2F52335485A294A4AFF46B77779A9CA86F3DA4D41A68021BB372BD2126BA2580CA629A19D7910344F6DD8C33E872E628E2231C56968018BB387976CEF885C1B065D587B4ACECEFBCDF6C5C71E5823504EF9EA3F12A0FF87103793A70D7F2DE4E66144981BC8B9C45EFCECE3ECF3A47EA72EE559047894D364357DE91DD1B2807D4AA5D7E2870CCCAC7C2C4E2099D510B4101DC4F2CD69ABD7F68AFC87B564BFFEF834E0E13D7FA69B8721D4035986EA9; GaiaAuthenticationValues=MjIvMTEvMjAyMyAxNDowNTo0NA==; GaiaRoteiro=340374; _BEAMER_FILTER_BY_URL_xcTSKWBK36426=true; chatWidgetWindowState2159A609ADD04A068939D64D16D9359F=false")
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

	fileName := fmt.Sprintf("%s/info.html", dir)
	err = os.WriteFile(fileName, body, 0644)
	if err != nil {
		log.Fatal("ERROR WRITING TO FILE ", err)
	}
	fmt.Printf("HTML content saved to %s\n", fileName)
}
