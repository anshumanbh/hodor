FROM golang:latest
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

RUN mkdir /api
WORKDIR /api
RUN mkdir tasks && mkdir worker

ADD tasks/ tasks/
ADD worker/ worker/

ENV GOBIN /go/bin
RUN mkdir -p /go/src/github.com/RichardKnop && cd /go/src/github.com/RichardKnop && git clone https://github.com/RichardKnop/machinery.git
RUN mkdir -p /go/src/k8s.io && cd /go/src/k8s.io && git clone https://github.com/kubernetes/client-go.git
RUN go install worker/worker.go

CMD ["worker"]
