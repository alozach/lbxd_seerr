package main

import (
	"github.com/alozach/lbxd_seerr/internal/lxbd"
	"github.com/alozach/lbxd_seerr/internal/scrapping"
)

func getWatchlist(s *scrapping.Scrapping, previousData []lxbd.Film) ([]lxbd.Film, error) {
	films, err := s.LxbdExtractFilms("/watchlist", previousData)
	if err != nil {
		return nil, err
	}

	VODFilms, err := s.LxbdExtractFilms("/watchlist/on/favorite-services", films)
	if err != nil {
		return nil, err
	}

	for i := range films {
		for _, vodf := range VODFilms {
			if films[i].Lid == vodf.Lid {
				films[i].VODAvailable = true
				break
			}
		}
	}
	return films, nil
}
