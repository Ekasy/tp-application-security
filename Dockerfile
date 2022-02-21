FROM golang:1.17 as builder

WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build ./cmd/main.go

FROM ubuntu:20.04

RUN apt-get -y update &&\
    apt-get -y install ca-certificates

WORKDIR /app
COPY --from=builder /app/main .

COPY . .

COPY /certs/ca.crt /usr/local/share/ca-certificates
COPY /certs/ca.crt /etc/ssl/certs/
RUN update-ca-certificates

CMD ./main
