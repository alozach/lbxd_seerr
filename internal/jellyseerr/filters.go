package jellyseerr

import (
	"fmt"
	"time"

	"github.com/alozach/lbxd_seerr/internal/lxbd"
)

type Filter struct {
	Name       string
	FilterFunc func(lxbd.Film) (bool, string)
}

var availableFilters = [...]Filter{
	{
		Name: "dry_run",
		FilterFunc: func(f lxbd.Film) (bool, string) {
			return false, ""
		}},
	{
		Name: "vod_not_available",
		FilterFunc: func(f lxbd.Film) (bool, string) {
			return !f.VODAvailable, ""
		}},
	{
		Name: "profitable",
		FilterFunc: func(f lxbd.Film) (bool, string) {
			details := fmt.Sprint("bud:", f.TmdbInfo.Budget, ", rev=", f.TmdbInfo.Revenue)
			return (f.TmdbInfo.Revenue > f.TmdbInfo.Budget && f.TmdbInfo.Budget > 0), details
		}},
	{
		Name: "released",
		FilterFunc: func(f lxbd.Film) (bool, string) {
			currentTime := time.Now()
			t, err := time.Parse("2006-01-02", f.TmdbInfo.ReleaseDate)
			if err != nil {
				return false, "failed to parse release date"
			}

			details := fmt.Sprint("release date: ", f.TmdbInfo.ReleaseDate)
			return t.Before(currentTime), details
		}},
}
