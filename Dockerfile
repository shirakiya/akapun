FROM golang:1.17.8

WORKDIR /go/src/

RUN apt-get update \
    && apt-get install -y --no-install-recommends zip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -L https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b /usr/local/bin v1.42.0
