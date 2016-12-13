FROM golang:latest
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

RUN mkdir /api
WORKDIR /api

ADD hodorapihandlers.go .
ADD hodorapilogger.go .
ADD hodorapimain.go .
ADD hodorapirouter.go .
ADD hodorapiroutes.go .

ENV GOBIN /go/bin
RUN mkdir -p /go/src/github.com/RichardKnop && cd /go/src/github.com/RichardKnop && git clone https://github.com/RichardKnop/machinery.git
RUN go get github.com/gorilla/mux
RUN go install hodorapimain.go hodorapihandlers.go hodorapilogger.go hodorapirouter.go hodorapiroutes.go

CMD ["hodorapimain", "hodorapihandlers", "hodorapilogger", "hodorapirouter", "hodorapiroutes"]
