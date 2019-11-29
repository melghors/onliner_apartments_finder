ARG HELM_VERSION

FROM golang:latest

RUN mkdir -p /go/src/onliner_apartments_finder

WORKDIR /go/src/onliner_apartments_finder

COPY config.json .
COPY main.go .

RUN go get -v
RUN go build .

ENV HELM_VERSION ${HELM_VERSION}

ENTRYPOINT ["/go/src/onliner_apartments_finder/onliner_apartments_finder"]

