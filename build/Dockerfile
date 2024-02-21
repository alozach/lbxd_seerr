
FROM selenium/standalone-chrome
MAINTAINER aloz96

WORKDIR /app
COPY . .

RUN mkdir -p logs

RUN sudo apt-get update
RUN sudo apt-get -y install cron

RUN crontab -l | { cat; echo "12 1 * * * /app/bin/watchlist_sync"; } | crontab -

ENTRYPOINT ["cron"]