FROM debian:stretch
COPY anime-sheduler settings.json ./
ENTRYPOINT ["./anime-sheduler"]