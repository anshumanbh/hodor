# hodor

## Features

* Built on Kubernetes with microservices pre-built as Docker images - no need to worry about OS, environments, languages, etc.
* Cluster Management of Docker containers handled by Kubernetes - no need to worry about scaling, load balancing, rolling updates, fault tolerance, etc.
* RESTFUL API to request tools to run - anybody can request a tool to run by hitting an API endpoint
* Asynchronous processing of jobs (running tools) - no need to wait for the results to come back. Get a UUID on submitting a job and retrieve the status of the job later using that UUID. Uses [machinery](https://github.com/RichardKnop/machinery) and Google PubSub for achieving this in two different places. Aim is to consolidate and just use 1 (most probably Google PubSub in the future).
* Extensible - ability to add more tools in the framework
* Aggregate output from multiple tools and ability to query them by leveraging Google Web Services such as Google Storage Cloud and Google BigQuery - make sense of the output from all the tools by doing analytics, machine learning, etc.

## Pre-requisites

* [Docker Toolbox](https://www.docker.com/products/docker-toolbox)
* [Kubectl](http://kubernetes.io/docs/user-guide/prereqs/)
* [Google ServiceAccount](https://cloud.google.com/compute/docs/access/service-accounts)
* [Minikube](http://kubernetes.io/docs/getting-started-guides/minikube/)
* In your Google Account, you will need:
    * the Project ID
    * to enable Google PubSub and create a topic and subscription for that topic. Grab the topic and subscription names.
    * to enable Google Storage Cloud and create a bucket to store the results. Grab the bucket name.
    * to enable Google BigQuery and create a Dataset. Grab the dataset name.
    * grab your Google Service Account credentials file and save it as `creds.json`. If you don't already have this file, you will need to create a service account credential (for Google App Engine) by navigating to https://console.developers.google.com/apis/credentials?project=<project-id>. Once you do that, save the .json file as creds.json somewhere on your local filesystem.

## Usage

Sending an API request (to the api/v1/runtool endpoint) via a CURL command to run a tool with multiple targets and options:

    curl -H "Content-Type: application/json" -X POST -d '{"Toolname":"<TOOLNAME>","Targets":["<IP1>", "<IP2>"],"Options":"<OPTIONS>"}' http://IP:PORT/api/v1/runtool

## Workflow

What the above command does is:
* API request sent to the server.
* Depending upon the number of targets (IPs above), jobs are created per target and are dropped in a queue.
* The jobs are picked up from the queue by multiple workers .
* The workers start the tool (Whatever tool is mentioned in the curl request above) against each target in separate Docker containers. This is possible because tools are built as Docker images.
* Results of the tool get uploaded to Google Storage Cloud (GSC).
* The results filename is dropped in a topic on Google PubSub.
* Subscription workers pick up the filename from the Google PubSub topic they are subscribed to.
* The subscription worker grabs that results file from GSC and uploads the data in Google BigQuery for further analysis, learning, etc.

## Architecture

![hodor](imgs/hodorarch.jpeg)

Hodor uses the following pre-built Docker images:
* [abhartiya/hodor_api:v1](https://hub.docker.com/r/abhartiya/hodor_api/)
* [abhartiya/machinery_worker:v1](https://hub.docker.com/r/abhartiya/machinery_worker/)
* [abhartiya/google_subscription:v1](https://hub.docker.com/r/abhartiya/google_subscription/)
* [abhartiya/tools_nmap](https://hub.docker.com/r/abhartiya/tools_nmap/)
* [rabbitmq](https://hub.docker.com/_/rabbitmq/)

The first 3 are no longer available because I took them down (because I am working on v2 of this project). You would need to re-build them using the Dockerfiles I have provided in each of the folders. So, in order to build the hodor api Docker image, navigate to hodor -> api, and type `docker build -t abhartiya/hodor_api:v1 .`.

For machinery worker, navigate to hodor -> api -> machinery and type `docker build -t abhartiya/machinery_worker:v1 .`.

For google subscription, navigate to hodor -> api -> googlesubscription and type `docker build -t abhartiya/google_subscription:v1 .`

If these Docker images failed to get built, it is most likely due to GO vendor problems. The Kubernetes client-go library changes frequently and its possible, the latest library doesn't work with the current code. If that's the case, please use something like [dep](https://github.com/golang/dep) to make sure you have the correct libraries.

## Documentation

* [Getting Started / Installation Tutorial](docs/getting-started.md)
* [ToDo List](docs/todo-list.md)

## Demo

Please watch this demo - https://youtu.be/rVFWttFoy5s