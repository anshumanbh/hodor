FROM golang:latest
MAINTAINER Anshuman Bhartiya anshuman.bhartiya@gmail.com

RUN mkdir /api
WORKDIR /api
RUN mkdir machinery && mkdir machinery/machinerytasks 

ADD machinerytasks/machinerytasks.go machinery/machinerytasks/machinerytasks.go
ADD machineryworker.go machinery/machineryworker.go

ENV GOBIN /go/bin

RUN go get cloud.google.com/go/pubsub
RUN go get cloud.google.com/go/storage
RUN go get github.com/lair-framework/go-nmap

RUN mkdir -p /go/src/github.com/RichardKnop && cd /go/src/github.com/RichardKnop && git clone https://github.com/RichardKnop/machinery.git
RUN mkdir -p /go/src/k8s.io && cd /go/src/k8s.io && git clone https://github.com/kubernetes/client-go.git
RUN go install machinery/machineryworker.go

CMD ["machineryworker"]
