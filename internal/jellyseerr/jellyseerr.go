package jellyseerr

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	c "github.com/alozach/lbxd_seerr/internal/config"
	"github.com/alozach/lbxd_seerr/internal/lxbd"
)

type Jellyseerr struct {
	ReqFilters       []Filter
	apiKey           string
	url              string
	requestedTMDbIds []int
	requestsLimit    int
	currNbRequests   int
}

type RequestStatus string

const (
	REQ_OK               RequestStatus = "REQ_OK"
	REQ_REACHED_LIMIT    RequestStatus = "REQ_REACHED_LIMIT"
	REQ_MISSING_DATA     RequestStatus = "MISSING_DATA"
	REQ_JELLYSEERR_ERROR RequestStatus = "JELLYSEERR_ERROR"
	REQ_ALREADY_OK       RequestStatus = "ALREADY_REQUESTED"
	REQ_FILTER_KO        RequestStatus = "FILTER_KO"
)

type Request struct {
	Film    lxbd.Film
	Status  RequestStatus
	Details string
}

const requestsFilename = "/app/data/last_requests.txt"

var js Jellyseerr

func Init(config c.JellyseerrConfig) {
	js = Jellyseerr{apiKey: config.ApiKey, url: config.BaseUrl + "/api/v1", requestsLimit: config.RequestsLimit}
}

func AddFilter(filterName string) {
	for _, f := range availableFilters {
		if f.Name == filterName {
			js.ReqFilters = append(js.ReqFilters, f)
			return
		}
	}
	log.Println("Invalid filter name: ", filterName)
}

func AddFilters(filterNames []string) {
	for _, name := range filterNames {
		AddFilter(name)
	}
}

func APICall(endpoint string, method string, body io.Reader) (*http.Response, error) {
	requestUrl := js.url + endpoint

	r, err := http.NewRequest(method, requestUrl, body)
	if err != nil {
		log.Printf("Failed to create %s request: %s", method, err)
		return nil, err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "application/json")
	r.Header.Add("X-Api-Key", js.apiKey)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		log.Printf("Failed to send %s request: %s", method, err)
		return nil, err
	}
	return res, nil
}

func RefreshRequestedTMDbIds() error {
	res, err := APICall("/request", "GET", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("Error getting Jellyseerr requests: got HTTP code %d", res.StatusCode)
		return errors.New("HTTP request failure")
	}

	var res_obj map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&res_obj)
	if err != nil {
		log.Println("Error parsing Jellyseer /request response: ", err)
		return err
	}

	var tmdbIds []int

	results := res_obj["results"].([]interface{})
	for _, result_if := range results {
		result := result_if.(map[string]interface{})
		media := result["media"].(map[string]interface{})
		tmdbId := int(media["tmdbId"].(float64))
		tmdbIds = append(tmdbIds, tmdbId)
	}

	js.requestedTMDbIds = tmdbIds
	return nil
}

func ResetRequestsCounter() {
	js.currNbRequests = 0
}

func CreateRequest(film lxbd.Film, refreshAlreadyRequested bool) Request {
	req := Request{Film: film}

	if film.TmdbInfo == nil {
		req.Status = REQ_MISSING_DATA
		return req
	}

	if js.requestedTMDbIds == nil || refreshAlreadyRequested {
		if err := RefreshRequestedTMDbIds(); err != nil {
			req.Status = REQ_JELLYSEERR_ERROR
			req.Details = err.Error()
			return req
		}
	}

	for _, tmdbId := range js.requestedTMDbIds {
		if tmdbId == film.TmdbInfo.ID {
			req.Status = REQ_ALREADY_OK
			return req
		}
	}

	for _, filter := range js.ReqFilters {
		filter_passed, details := filter.FilterFunc(film)
		if !filter_passed {
			retDetails := filter.Name
			if details != "" {
				retDetails += ": " + details
			}
			req.Status = REQ_FILTER_KO
			req.Details = retDetails
			return req
		}
	}

	if js.requestsLimit > 0 && js.currNbRequests >= js.requestsLimit {
		req.Status = REQ_REACHED_LIMIT
		return req
	}

	body, _ := json.Marshal(map[string]interface{}{"mediaType": "movie", "mediaId": film.TmdbId, "userId": 2})

	res, err := APICall("/request", http.MethodPost, bytes.NewBuffer(body))
	if err != nil {
		req.Status = REQ_JELLYSEERR_ERROR
		req.Details = err.Error()
		return req
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		req.Status = REQ_JELLYSEERR_ERROR
		req.Details = fmt.Sprint("Got HTTP code", res.StatusCode)
		return req
	}

	js.requestedTMDbIds = append(js.requestedTMDbIds, film.TmdbId)
	req.Status = REQ_OK
	js.currNbRequests++
	return req
}

func SaveRequests(requests []Request) error {
	requestsFile, err := os.Create(requestsFilename)
	if err != nil {
		log.Println("Failed to create request data: ", err)
		return err
	}
	defer requestsFile.Close()

	writer := csv.NewWriter(requestsFile)
	defer writer.Flush()

	headers := []string{"tmdbId", "name", "status", "details"}

	writer.Write(headers)
	for _, req := range requests {
		row := []string{strconv.Itoa(req.Film.TmdbId), req.Film.TmdbInfo.Title, string(req.Status), req.Details}
		writer.Write(row)
	}
	return nil
}

func GetLastRequets() ([]byte, error) {
	file, err := os.Open(requestsFilename)

	if err != nil {
		log.Println("Error while reading the file", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	// Read the CSV data rows
	var data []map[string]interface{}
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		m := make(map[string]interface{})
		for i, val := range row {
			valint, err := strconv.ParseInt(val, 10, 0)
			if err == nil {
				m[headers[i]] = valint
				continue
			}

			m[headers[i]] = val
		}
		data = append(data, m)
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return b, nil
}
