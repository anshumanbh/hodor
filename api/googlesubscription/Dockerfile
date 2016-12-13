FROM golang:latest
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

ADD googlesubscription.go .

ENV GOBIN /go/bin
RUN go get cloud.google.com/go/pubsub
RUN go get cloud.google.com/go/bigquery
RUN go install googlesubscription.go

CMD ["googlesubscription"]
