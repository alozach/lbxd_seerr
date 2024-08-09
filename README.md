LbxdSeer aims to link a [Jellyseer](https://github.com/Fallenbagel/jellyseerr) instance to a [Letterboxd](https://letterboxd.com/) account, and do some stuff between these 2

## Current features

### Tasks

Tasks ran periodically if enabled:
* `dl_watchlist` : Scrap your Letterboxd watchlist and create a Jellyseer download request for each movie not in your Jellyseer list yet, filtering them according to your config (see below)

### API

Provides an API to get info about LbxdSeer actions. Implemented endpoints:
* `GET /requests` : Get the status of last requests sent to the Jellyseer instance


## Configuration

A config file is expected to be found in `/config/lbxd_seerr.yaml` (in the Docker containers)

```
lxbd:
  username: string
  password: string

tmdb:
  api_key: string

jellyseerr:
  api_key: string
  base_url: string
  requests_limit: int
  filters:
    - released
    - vod_not_available
    - profitable
    - dry_run
tasks:
  dl_watchlist: cron expression (e.g. 0 0 * * *)
```

* `lbxd` : Letterboxd username / password
* `tmdb`:
    * `api_key` : [TMDB](https://www.themoviedb.org/?language=fr) API key
* `jellyseer`:
    * `api_key` : Jellyseer API key
    * `base_url`: url of the Jellyseer instance
    * `request_limit`: Max number of requests to send to Jellyseer in one `dl_watchlist` iteration
    * `filters`: Do not send a request for movies not passing these filters. Comment a filter to disable it
        * `released`: Movie has to be released in theaters
        * `vod_not_available`: Movie must not be available on any streaming service listed in Letterboxd "Favorite services"
        * `profitable`: Movie revenue has to be higher then its budget (info taken from TMDB)
        * `dry_run`: No movie will be requested with this filter enabled
    * `tasks`: List of tasks to be run periodically. Comment a task to disable it
        * `dl_watchlist`: See description above


## Known limitations

*  It'd be nice not to have to log in to the Letterboxd account. This is needed to be able to access the user watchlist in a single page. Not logging in to the account shows a watchlist in pagination way, leading to scrapping difficulties (don't remember which ones ¯\_(ツ)_/¯)
