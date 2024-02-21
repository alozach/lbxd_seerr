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

var js Jellyseerr

func Init(apiKey string, baseUrl string) error {
	js = Jellyseerr{apiKey: apiKey, url: baseUrl + "/api/v1"}
	if err := RefreshRequestedTMDbIds(); err != nil {
		return err
	}
	return nil
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

func CreateRequest(film lxbd.Film) bool {
	if film.TmdbInfo == nil {
		log.Println("Missing TMDb info for film with lid ", film.Lid)
		return false
	}

	for _, tmdbId := range js.requestedTMDbIds {
		if tmdbId == film.TmdbInfo.ID {
			log.Printf("Already requested \"%s\" (%d)", film.TmdbInfo.Title, film.TmdbId)
			return false
		}
	}

	for _, filter := range js.ReqFilters {
		filter_passed, details := filter.FilterFunc(film)
		if !filter_passed {
			errLog := fmt.Sprint(filter.Name, " filter did not pass for film ", film.TmdbInfo.Title, " (", film.TmdbId, ")")
			if details != "" {
				errLog += ": " + details
			}
			log.Println(errLog)
			return false
		}
	}

	body, _ := json.Marshal(map[string]interface{}{"mediaType": "movie", "mediaId": film.TmdbId, "userId": 2})

	res, err := APICall("/request", http.MethodPost, bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		log.Printf("Error creating request for  \"%s\" (TMDb id %d): got HTTP code %d", film.TmdbInfo.Title, film.TmdbId, res.StatusCode)
		return false
	}

	log.Printf("Successfully created request for \"%s\" (TMDb id %d)", film.TmdbInfo.Title, film.TmdbId)
	js.requestedTMDbIds = append(js.requestedTMDbIds, film.TmdbId)
	return true
}
