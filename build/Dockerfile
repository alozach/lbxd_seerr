
FROM selenium/standalone-chrome
MAINTAINER aloz96

USER 0
WORKDIR /app

COPY bin bin/
COPY resources resources/

RUN sudo chown -R seluser:seluser /app

ENTRYPOINT ["/app/bin/watchlist_sync"]