FROM golang:1.13.5-alpine3.11 AS godev

RUN apk update && apk add --no-cache ca-certificates && apk upgrade && apk add git

WORKDIR /rss-sync

COPY . .
ENV GO111MODULE=on
ENV GOSUMDB=off
ENV GOPROXY=direct

RUN go build -o rss-sync .

FROM alpine:3.11

RUN apk update && apk add --no-cache ca-certificates && apk upgrade

COPY --from=godev ./rss-sync/rss-sync /rss-sync