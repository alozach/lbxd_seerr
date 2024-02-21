package jellyseerr

import (
	"fmt"

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
			return f.TmdbInfo.Revenue >= f.TmdbInfo.Budget, details
		}},
}
