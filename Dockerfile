# syntax=docker/dockerfile:1.2
#######################################################################################################################
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./ 
RUN go mod download

COPY cmd/  cmd
COPY rpc/ rpc 
COPY config/ config 

RUN go mod tidy
RUN CGO_ENABLED=0 go build -o storage_server -ldflags "-w -s" cmd/rpc_storage_server/*
RUN CGO_ENABLED=0 go build -o http_server  -ldflags "-w -s" cmd/http_server/*

#######################################################################################################################

FROM alpine:3.10 AS http

COPY --from=builder /app/http_server ./ 
COPY templates/ templates/
ENTRYPOINT ["./http_server"]

#######################################################################################################################

FROM alpine:3.10 AS storage

COPY --from=builder /app/storage_server ./ 
ENTRYPOINT ["./storage_server"]

#######################################################################################################################
