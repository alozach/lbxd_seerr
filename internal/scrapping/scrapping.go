package scrapping

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	"github.com/ryanbradynd05/go-tmdb"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"

	"github.com/alozach/lbxd_seerr/internal/lxbd"
)

type Scrapping struct {
	Collector    *colly.Collector
	Driver       selenium.WebDriver
	service      *selenium.Service
	lxbdUsername string
	tmdbAPI      *tmdb.TMDb
}

const lxbdBaseUrl string = "https://letterboxd.com"

func (scrapping *Scrapping) initColly() {
	scrapping.Collector = colly.NewCollector(
		colly.Async(true), // Enable asynchronous requests
	)
	scrapping.Collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 6,
	})
}

func (scrapping *Scrapping) initSelenium() {
	var err error

	scrapping.service, err = selenium.NewChromeDriverService("./resources/chromedriver", 4444)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"--headless",
		"--no-sandbox",
		"--user-agent=Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	}})

	scrapping.Driver, err = selenium.NewRemote(caps, "")
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = scrapping.Driver.MaximizeWindow("")
	if err != nil {
		log.Fatal("Error maximizing window: ", err)
	}
}

func Init(tmdbApiKey string) *Scrapping {
	scrapping := &Scrapping{}
	scrapping.initColly()
	scrapping.initSelenium()

	config := tmdb.Config{
		APIKey:   tmdbApiKey,
		Proxies:  nil,
		UseProxy: false,
	}
	scrapping.tmdbAPI = tmdb.Init(config)

	return scrapping
}

func Deinit(s *Scrapping) {
	s.service.Stop()
}

func (scrapping *Scrapping) LxbdAcceptCookies() error {
	err := scrapping.Driver.Get(lxbdBaseUrl)
	if err != nil {
		log.Println("Error getting page: ", err)
		return err
	}

	err = scrapping.Driver.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		elem, _ := driver.FindElement(selenium.ByClassName, "fc-cta-consent")
		if elem != nil {
			return elem.IsDisplayed()
		}
		return false, nil
	}, 15*time.Second)

	if err != nil {
		log.Println("Timeout waiting for consent popup")
		return nil
	}

	consentButton, err := scrapping.Driver.FindElement(selenium.ByClassName, "fc-cta-consent")
	if err != nil {
		log.Println("Failed to get Consent div: ", err)
		return err
	}
	consentButton.Click()

	log.Println("Cookies accepted")
	return nil
}

func (scrapping *Scrapping) LxbdLogIn(username string, password string) error {
	err := scrapping.Driver.Get(lxbdBaseUrl)
	if err != nil {
		log.Println("Error getting page: ", err)
		return err
	}

	signInButton, err := scrapping.Driver.FindElement(selenium.ByClassName, "sign-in-menu")
	if err != nil {
		log.Println("Failed to get sign in button: ", err)
		return err
	}
	signInButton.Click()

	formElement, err := scrapping.Driver.FindElement(selenium.ByID, "signin")
	if err != nil {
		log.Println("Failed to get sign in form: ", err)
		return err
	}

	input, err := formElement.FindElement(selenium.ByName, "username")
	if err != nil {
		log.Println("Failed to get username field: ", err)
		return err
	}
	input.SendKeys(username)

	input, err = formElement.FindElement(selenium.ByName, "password")
	if err != nil {
		log.Println("Failed to get username field: ", err)
		return err
	}
	input.SendKeys(password)

	err = formElement.Submit()
	if err != nil {
		log.Println("Failed to submit login form: ", err)
		return err
	}

	err = scrapping.Driver.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		if _, err := driver.GetCookie("letterboxd.signed.in.as"); err != nil {
			return false, nil
		}
		return true, nil
	}, 10*time.Second)

	if err != nil {
		log.Println("Timeout waiting logged in cookie")
		return err
	}

	log.Println("Successfully logged in")

	scrapping.lxbdUsername = username
	return nil
}

func (s *Scrapping) fetchTMDbId(film *lxbd.Film) error {
	s.Collector.OnHTML("body", func(e *colly.HTMLElement) {
		tmdbIdStr := e.Attr("data-tmdb-id")
		tmdbId, err := strconv.Atoi(tmdbIdStr)
		if err != nil {
			log.Printf("Could not convert \"%s\" to TMDb id", tmdbIdStr)
		} else {
			film.TmdbId = tmdbId
			log.Printf("Fetched tmdbId %d for lid %d", tmdbId, film.Lid)
		}
	})

	url := lxbdBaseUrl + film.LxbdEndpoint
	err := s.Collector.Visit(url)
	defer func() {
		s.Collector.Wait()
		s.Collector.OnHTMLDetach("body")
	}()

	if err != nil {
		log.Printf("Failed to visit %s: %s", url, err)
		return err
	}

	return nil
}

func (s *Scrapping) fetchTMDbInfo(film *lxbd.Film) error {
	err := s.fetchTMDbId(film)
	if err != nil {
		return err
	}
	film.TmdbInfo, err = s.tmdbAPI.GetMovieInfo(film.TmdbId, nil)
	if err != nil {
		log.Printf("Failed to get TMDb info for lid %d: %s", film.Lid, err)
		film.TmdbInfo = nil
		return errors.New("failed to get TMDb info")
	}
	return nil
}

func (scrapping *Scrapping) LxbdExtractFilms(endpoint string, previousData []lxbd.Film) ([]lxbd.Film, error) {
	var films []lxbd.Film

	url := lxbdBaseUrl + "/" + scrapping.lxbdUsername + endpoint
	err := scrapping.Driver.Get(url)
	if err != nil {
		log.Println("Error getting page: ", err)
		return nil, err
	}

	// ids can take some time to appear
	time.Sleep(1 * time.Second)

	filmsDiv, err := scrapping.Driver.FindElements(selenium.ByClassName, "film-poster")
	if err != nil {
		log.Println("Failed to get films: ", err)
		return nil, err
	}

	for _, div := range filmsDiv {
		name, err := div.GetAttribute("data-film-slug")
		if err != nil {
			log.Println("Failed to get film name: ", err)
		}

		lidStr, err := div.GetAttribute("data-film-id")
		if err != nil {
			log.Printf("Failed to get id of film \"%s\": %s", name, err)
			continue
		}
		lid, err := strconv.Atoi(lidStr)
		if err != nil {
			log.Printf("Could not convert \"%s\" to LID", lidStr)
			continue
		}

		link, err := div.GetAttribute("data-film-link")
		if err != nil {
			link, err = div.GetAttribute("data-target-link")
			if err != nil {
				log.Printf("Failed to get link of film \"%s\" (%d): %s", name, lid, err)
				continue
			}
		}

		var film *lxbd.Film
		for _, prev := range previousData {
			if lid == prev.Lid {
				film = &prev
				log.Printf("Using previously fetched tmdbInfo for lid %d", lid)
			}
		}

		if film == nil {
			film = &lxbd.Film{Lid: lid, LxbdEndpoint: link}
		}

		if film.TmdbInfo == nil {
			if err := scrapping.fetchTMDbInfo(film); err != nil {
				continue
			}
		}

		films = append(films, *film)

	}

	return films, nil
}
