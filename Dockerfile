FROM golang:latest
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

RUN mkdir /api
WORKDIR /api

ADD api/*.go ./

ENV GOBIN /go/bin
RUN mkdir -p /go/src/github.com/RichardKnop && cd /go/src/github.com/RichardKnop && git clone https://github.com/RichardKnop/machinery.git
RUN go get github.com/gorilla/mux
RUN go install main.go handlers.go logger.go router.go routes.go

CMD ["main", "handlers", "logger", "router", "routes"]