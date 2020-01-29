#!/bin/bash

dep ensure
go build -o anime-sheduler
docker build -t ivantimofeev/anime-sheduler .
docker push ivantimofeev/anime-sheduler
