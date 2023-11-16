FROM golang:1.19

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY utils ./utils

RUN go build -o fetch

EXPOSE 8080