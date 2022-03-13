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

COPY /gen_certificate/ca.crt /usr/local/share/ca-certificates
COPY /gen_certificate/ca.crt /etc/ssl/certs/

RUN ./gen_certificate/gen_ca.sh &&\
    update-ca-certificates

CMD ./main
