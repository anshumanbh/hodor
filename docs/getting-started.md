## Getting Started / Installation Tutorial

This tutorial will walk you through the installation and setup of hodor (running on a single node Kubernetes cluster) locally using [minikube](http://kubernetes.io/docs/getting-started-guides/minikube/). It will also walk you through the setup of hodor locally for doing development.

### Local Installation (with in-cluster config)

* Install docker-machine, kubectl and minikube as mentioned in the pre-requisites in [README](../README.md)

The versions I have tested this with are:
    * docker-machine version 0.8.2, build e18a919
    * kubectl Client Version: version.Info{Major:"1", Minor:"4", GitVersion:"v1.4.6", GitCommit:"e569a27d02001e343cb68086bc06d47804f62af6", GitTreeState:"clean", BuildDate:"2016-11-12T05:22:15Z", GoVersion:"go1.7.1", Compiler:"gc", Platform:"darwin/amd64"}
    * kubectl Server Version: version.Info{Major:"1", Minor:"4", GitVersion:"v1.4.3", GitCommit:"4957b090e9a4f6a68b4a40375408fdc74a212260", GitTreeState:"clean", BuildDate:"1970-01-01T00:00:00Z", GoVersion:"go1.7.1", Compiler:"gc", Platform:"linux/amd64"}
    * minikube version: v0.13.1

* Git Clone this repository

* Type `minikube start` to start a single node Kubernetes cluster on your local machine

* Type `minikube dashboard` to start the Kubernetes dashboard in your browser

* Type `eval $(minikube docker-env)` to setup your docker config to point to minikube's docker environment

* If you look at the `.yaml` files in the `hodor/config` directory, you will notice that I am using some environment variables for `PROJECT_ID`, `PUBSUB_TOPICNAME`, `BUCKET_NAME`, `SUBSCRIPTION_NAME`, `DATASET_NAME`, `TABLE_NAME` and `GOOGLE_APPLICATION_CREDENTIALS` in some of those files.

* Leave the `GOOGLE_APPLICATION_CREDENTIALS` and `TABLE_NAME` environment variables as is.

* Replace `PROJECT_ID`, `PUBSUB_TOPICNAME`, `BUCKET_NAME`, `SUBSCRIPTION_NAME`, `DATASET_NAME` values with the values corresponding to your Google Service Account that you needed as part of the Pre-requisites mentioned in [README](../README.md).

* Create a Kubernetes secret from your Google service account credentials file (mentioned in the Pre-requisites as well) like this:
`kubectl create secret generic googlesecret --from-file=/path/to/your/creds.json`

* Make sure you can see it by typing `kubectl get secret googlesecret -o yaml`

* At this point, you should be set to start your deployments in the Kubernetes cluster.

* There are 6 yaml files right now in the `hodor/config` directory:
    * googlesubscription-deployment.yaml (deploys 1 replica for the Google Subscription worker)
    * hodor/api/hodorapi-deployment.yaml (deploys 2 replicas for the API server running on port 3636)
    * hodor/api/hodorapi-service.yaml (exposes and load balances the API server)
    * hodor/api/worker/machineryworker-deployment.yaml (deploys 3 replicas for the machinery workers that consumes the message from the rabbitmq broker)
    * hodor/rabbitmq/rabbitmq-deployment.yaml (deploys 1 replica for the rabbitmq broker running on port 5672)
    * hodor/rabbitmq/rabbitmq-service.yaml (exposes the rabbitmq broker to the API servers as rabbitmq-service to talk to and submit jobs/tasks)

* Navigate to the `hodor` directory and run `kubectl create -f config/ --recursive`.

* Once you do this, your cluster should be up and running. 

* You can also navigate to the `minikube dashboard` to see all the Pods deployed there.

* Next, find out the IP of your Minikube docker-env by typing `minikube docker-env` and capturing the IP under `DOCKER_HOST`.

* Next, find the port the API service is exposed to by navigating to `Services` in the minikube dashboard. It should be the one under port 80. Minikube doesn't loadbalance properly since its running locally and not remotely. So, the port `80` that we really want to expose wouldn't work. The other port works though so use that port below. 

* Finally, send a curl request (also mentioned in the USAGE section in README) to the API service such as `curl -H "Content-Type: application/json" -X POST -d '{"Toolname":"<TOOLNAME>","Targets":["<IP1>", "<IP2>"],"Options":"<OPTIONS>"}' http://IP:PORT/api/v1/runtool`. You should receive 2 Task IDs in the response. Navigate to the worker pods and look at their logs. Notice that the workers picked up the tasks and completed it. They then sent the filename fo Google PubSub and stored the scan result file (after converting to CSV) to Google Storage Cloud. 

* The Subscription worker consumed the filename and ran another job taking the CSV data from Google Storage Cloud and uploading it to Google BigQuery. That's it!


### Local Development (with out-cluster config) 

You will notice that it is not feasible to build the entire minikube cluster locally every time you do a small change in the codebase. That would involve deleting all deployments, pods, jobs, services, rebuilding all Docker images, redeploying them, etc.
In order to prevent all that work, we can actually develop locally by testing and making sure everything runs fine and its okay to deploy it in a cluster.
In order to do that, we need to develop with an out-cluster config. For that, we need to make the following changes:

* In hodor/api/machinery/machinerytasks/machinerytasks.go file, comment the first 2 lines of the Runtool function and uncomment out the next 2 lines. This basically changes the in-cluster config to out-cluster config for Kubernetes. Notice that instead of letting Minikube use its config, we are now specifying a config file to manage the pods out of cluster. 
   
* In hodor/api/machinery/machineryworker.go file, search for "rabbitmq-service" and replace all occurrences by "localhost". We do this because we will be starting the RabbitMQ service on the localhost out of cluster for the ease of testing.
    
* In hodor/api/hodorapihandlers.go file, do the same - Search for "rabbitmq-service" and replace all occurrences by "localhost" for similar reasons as stated above.
    
* Do whatever change you want in the codebase. No need to build the Docker images. You can start 1 instance of all the services individually as mentioned below:
    
* In one terminal, start the rabbitmq service by typing "rabbitmq-server". This will start the RabbitMQ service.
    
* In another terminal, navigate to the hodor/api/machinery directory. You need the following environment variables for this terminal:
        - export PROJECT_ID=<PROJECTID>
        - export PUBSUB_TOPICNAME=<PUBSUBTOPICNAME>
        - export BUCKET_NAME=<BUCKETNAME>
        - export GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/creds.json
        Then, run "go run machineryworker.go". This will start the workers to pick up tasks from the RabbitMQ broker. 
    
* In another terminal, navigate to hodor/api directory and run "go run *.go". This will start all the components of the API server. 
    
* In another terminal, navigate to hodor/api/googlesubscription directory. You need the following environment variables for this terminal:
        - export PROJECT_ID=<PROJECTID>
        - export SUBSCRIPTION_NAME=<SUBSCRIPTIONNAME>
        - export DATASET_NAME=<DATASETNAME>
        - export TABLE_NAME=<TABLENAME>
        - export BUCKET_NAME=<BUCKETNAME>
        - export GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/creds.json
        Then, run "go run googlesubscription.go". This will start the Subscription workers. 

* The infrastructue is up and running now. You can start sending API requests to the API server running locally on port 3636 such as "curl -H "Content-Type: application/json" -X POST -d '{"Toolname":"<TOOLNAME>","Targets":["<IP1>", "<IP2>"],"Options":"<OPTIONS>"}' http://localhost:3636/api/v1/runtool". Rest is similar as above. 