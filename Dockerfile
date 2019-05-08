FROM golang:1.10.0-alpine

# add git
RUN apk update \ 
    && apk add \
        bash \
        curl \
        git

# add dep
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep

# COPY Gopkg.lock Gopkg.toml /go/src/github.com/puppetlabs/cloud-discovery/vault/
COPY . /go/src/github.com/puppetlabs/cloud-discovery/vault/

WORKDIR /go/src/github.com/puppetlabs/cloud-discovery/vault

RUN dep ensure

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o /vault_client_bin cmd/vault_client/main.go