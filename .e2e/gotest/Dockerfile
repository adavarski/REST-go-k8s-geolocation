FROM golang:1.16-alpine

ADD .e2e/gotest/go.* /go/src/

WORKDIR /go/src/

RUN go mod download

COPY . /go/src/

ARG BASE_URL=http://geolocation-go
ARG REDIS_CONNECTION_STRING=redis://redis:6379
ENV CGO_ENABLED 0

WORKDIR /go/src/.e2e/gotest

ARG BASE_URL=http://geolocation-go
ARG REDIS_CONNECTION_STRING=redis://redis:6379
ARG PPROF_URL=http://geolocation-go:6060/debug/pprof/
ENV BASE_URL ${BASE_URL}
ENV REDIS_CONNECTION_STRING ${REDIS_CONNECTION_STRING}
ENV PPROF_URL ${PPROF_URL}