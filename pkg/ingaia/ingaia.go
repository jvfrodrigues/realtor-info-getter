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

	"github.com/jvfrodrigues/realtor-info-getter/pkg/shared"
)

var BEARER_TOKEN string
var COOKIES = "showNewMenu=false; SignInNoPlans=12/09/2023 19:43:59; _BEAMER_USER_ID_xcTSKWBK36426=875111a9-a7e7-4547-a241-e94ece7b66e3; _BEAMER_FIRST_VISIT_xcTSKWBK36426=2023-05-14T20:50:07.675Z; _BEAMER_LAST_PUSH_PROMPT_INTERACTION_xcTSKWBK36426=1701793983655; _BEAMER_BOOSTED_ANNOUNCEMENT_DATE_xcTSKWBK36426=2023-11-12T15:20:10.961Z; _BEAMER_DATE_xcTSKWBK36426=2023-10-24T19:27:47.193Z; _BEAMER_LAST_UPDATE_xcTSKWBK36426=1702161844118; __goc_session__=gmuctzfioukjldcbcuguhcatoozvuhgq; ASP.NET_SessionId=s0o51bg1ytlbpvwhvfs5mecs; B=F54C6885340A19F84B3414A75BDEE8C3079EBC1A; GaiaAuthenticationLoginTime=12/09/2023 19:43:59; teste_https=value_teste_https; enterprises_url=https://enterprises.ingaia.com.br; signalr_url=https://chat.ingaia.com.br; chat_api_url=https://api-chat.ingaia.com.br; contacts_api_url=https://contacts-chat.ingaia.com.br; gaiacore_url=https://imob.gaiacore.com.br/; matomo=3; environment=Production; listings_url=https://listings.ingaia.com.br; accounts_url=https://accounts.ingaia.com.br; goals_url=https://imob-goals.ingaia.com.br; leads_url=https://leads.ingaia.com.br; clients_radar_url=https://radar.ingaia.com.br; customers_url=https://customers.ingaia.com.br; customer_search_url=https://customer-search.ingaia.com.br; kernel_url=https://kernel.ingaia.com.br; locations_url=https://locations.ingaia.com.br; integration_rd_url=https://integration-rdstation.ingaia.com.br/; rd_url=https://api.rd.services/auth/dialog?client_id=fe274f90-9def-4a76-a7b3-3edeafdc6e11; pipe_leads_url=https://pipe-leads.ingaia.com.br/; notifications_url=https://api-notifications.ingaia.com.br; notifications_socket_url=wss://notifications.ingaiaimobApiNotificationscom.br; page_dashboard_loan_url=https://credito.valuegaia.com.br; cms_authorize_page=https://sites-cms.kenlo.io/authorize; mf_onboarding_dashboard_loan_url=https://mf-broker-onboardings-z6rnix554q-uc.a.run.app; mf_kenlo_expert_presentation_url=https://mf-kenlo-expert-presentation-z6rnix554q-uc.a.run.app; config_lead_rules_mf_url=https://config-lead-rules.imob.stg.kenlo.io; home_equity_bff=https://us-central1-kenlo-kash.cloudfunctions.net/ms-bff-opportunities-prod-main; gaiacore_token=4dd7f43506c6bcfae512adbad92e3953; accounts_token=eyJhbGciOiJSUzI1NiIsImtpZCI6IjdDRUFBRkRFRDFBODEwRjBCNzVFRjE0OTBBNjNGNTA1MTA2Mzg5MjYiLCJ0eXAiOiJKV1QiLCJ4NXQiOiJmT3F2M3RHb0VQQzNYdkZKQ21QMUJSQmppU1kifQ.eyJuYmYiOjE3MDIxNjE4NDAsImV4cCI6MTcwMjI0ODI0MCwiaXNzIjoiaHR0cDovL2FjY291bnRzLmluZ2FpYS5jb20uYnIiLCJhdWQiOlsiaHR0cDovL2FjY291bnRzLmluZ2FpYS5jb20uYnIvcmVzb3VyY2VzIiwiQVBJIl0sImNsaWVudF9pZCI6InZhbHVlZ2FpYSIsInN1YiI6ImFsZnJlZG9Abm92YWNhc2FyYW8uY29tLmJyIiwiYXV0aF90aW1lIjoxNzAyMTYxODQwLCJpZHAiOiJsb2NhbCIsImlkIjoiMzQwMzc0IiwiaHR0cDovL3NjaGVtYXMubWljcm9zb2Z0LmNvbS93cy8yMDA4LzA2L2lkZW50aXR5L2NsYWltcy9yb2xlIjoiQ29ycmV0b3IiLCJuYW1lIjoiQWxmcmVkbyBMb3VyZW7Dp28gUm9kcmlndWVzIiwicGhvdG8iOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vSW0xNkNHaW1GNTFZaGRfMlhaOG9hdmFwWHZtcmF0SW1XX08zZG5XZ01ObXJoMXdhVDBCcFc5VHFuZERxQWxVM3ItMHQ2M3kxOVJBaU1JSWF6OXlaMmV5b3NQcTR5cy1LdmxBNkQzUG1xME9QY0ZyWj13MTAyNC1oNzY4IiwiY29tcGFueV9pZCI6IjM3NTA2IiwiY29tcGFueV9uYW1lIjoiTm92YSBDYXNhcsOjbyBJbcOzdmVpcyIsIm9mZmljZV9pZCI6IjYwMjU0IiwidGVhbV9pZCI6IjAiLCJwcm9maWxlX2lkIjoiMyIsIm5ldHdvcmtzIjoiW10iLCJjb250cmFjdF9zbCI6Im5vdmFfY2FzYXJhb19pbW92ZWlzX2x0ZGFtZTAxIiwic2NvcGUiOlsiY2hhdCIsImRhdGFpbW9iIiwiaW50ZWdyYXRpb24iLCJsaXN0aW5ncyJdLCJhbXIiOlsiY3VzdG9tQXV0aGVudGljYXRpb24iXX0.3L6_YXltSvSphuoXuOVqYAbyTTX_RANOWdivlv8fWTqE8kwZm1XdgZGGu06wtpnLKdmrhdBxYm0kTb-7HWktFWMERuD44HWbKuN7_gEaNuwSOmFlCOtSyyNtuIgIe65RDjA9gN71i_cX4Hrzkz9RTXQwH5QVyg_vq1fXx1bR5eHjSF5udW8-c0H8HTqurFAjDrO1_hhwM86m_LhJkMuDHe3SEmcGlJDTVOlmUai2nwDLA2_Izc6h6rdo_5SwFfaTLigvaMIkJeQVwNzafAtBa92K_d5vw8ERFRKAO8j_YAeQhX2jApS1iQ2YHACBHDA6Hy5SZksVr6L7_Bwzyhuy9A; GaiaAuthentication=7E888A0E8A99C02E841EF86DA3B3CD81B9D0583D4571C406CEDFD3DAA97DF12B84E7EA2EB2702515A17E87048D1FDBFFC4C7C688E160044F22CFD628F1483A7CFD682D3CCADE0030AE5AD2EA414581AE20B99FA01F2347C10DF71805EFFE80DB8DB15C8815D939526B54CAB8726CB1C9BB5939A695A9B77016AD71F33562B6B8E7A9CFC648CD33F77D9722FC74EC262179910427604195A63FCA7EC6A67A8E2B10D421CD2E74C18DD998520DA68445FA8E7C5B96EA7DC5345416728A1C78C9CBFC65F25677A134A1C637BE7D772962564AADFBFA19D7B9339856ADEDF0FEEE61B57BC1903FD39BBB07B76AA8C8D7EA3061A802DAFB00AF4E194D53F71DB13E910CFB8689723523A1F00FABDC9F5900DB; GaiaAuthenticationValues=MDkvMTIvMjAyMyAxOTo0NDowMw==; chatWidgetWindowState2159A609ADD04A068939D64D16D9359F=false; GaiaRoteiro=340374; _BEAMER_FILTER_BY_URL_xcTSKWBK36426=true"

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

func GetIngaia() {
	BEARER_TOKEN = shared.StringPrompt("BEARER TOKEN:")
	if BEARER_TOKEN == "" {
		log.Panicln("Bearer cant be nil")
	}
	saveFolder := os.Getenv("FOLDER_PATH_AUTO")
	if saveFolder == "" {
		log.Panicln("Env var for path to folder not defined")
	}
	start := time.Now()
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
	req.Header.Add("Authorization", BEARER_TOKEN)
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
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://imob.valuegaia.com.br/admin/modules-react/")
	req.Header.Add("Cookie", COOKIES)
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
	err = os.WriteFile(fileName, body, 0644)
	if err != nil {
		log.Fatal("ERROR WRITING TO FILE ", err)
	}
	fmt.Printf("HTML content saved to %s\n", fileName)
}
