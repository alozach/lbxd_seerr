package jellyseerr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alozach/lbxd_seerr/internal/lxbd"
)

type Jellyseerr struct {
	ReqFilters       []Filter
	apiKey           string
	url              string
	requestedTMDbIds []int
}

type RequestStatus string

const (
	REQ_OK               RequestStatus = "REQ_OK"
	REQ_MISSING_DATA     RequestStatus = "MISSING_DATA"
	REQ_JELLYSEERR_ERROR RequestStatus = "JELLYSEERR_ERROR"
	REQ_ALREADY_OK       RequestStatus = "ALREADY_REQUESTED"
	REQ_FILTER_KO        RequestStatus = "FILTER_KO"
)

type Request struct {
	Status  RequestStatus
	Details string
}

var js Jellyseerr

func Init(apiKey string, baseUrl string) {
	js = Jellyseerr{apiKey: apiKey, url: baseUrl + "/api/v1"}
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

func CreateRequest(film lxbd.Film, refreshAlreadyRequested bool) Request {
	if js.requestedTMDbIds == nil || refreshAlreadyRequested {
		if err := RefreshRequestedTMDbIds(); err != nil {
			return Request{Status: REQ_JELLYSEERR_ERROR, Details: err.Error()}
		}
	}

	if film.TmdbInfo == nil {
		return Request{Status: REQ_MISSING_DATA}
	}

	for _, tmdbId := range js.requestedTMDbIds {
		if tmdbId == film.TmdbInfo.ID {
			return Request{Status: REQ_ALREADY_OK}
		}
	}

	for _, filter := range js.ReqFilters {
		filter_passed, details := filter.FilterFunc(film)
		if !filter_passed {
			retDetails := filter.Name
			if details != "" {
				retDetails += ": " + details
			}
			return Request{Status: REQ_FILTER_KO, Details: retDetails}
		}
	}

	body, _ := json.Marshal(map[string]interface{}{"mediaType": "movie", "mediaId": film.TmdbId, "userId": 2})

	res, err := APICall("/request", http.MethodPost, bytes.NewBuffer(body))
	if err != nil {
		return Request{Status: REQ_JELLYSEERR_ERROR, Details: err.Error()}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return Request{Status: REQ_JELLYSEERR_ERROR, Details: fmt.Sprint("Got HTTP code", res.StatusCode)}
	}

	js.requestedTMDbIds = append(js.requestedTMDbIds, film.TmdbId)
	return Request{Status: REQ_OK}
}
