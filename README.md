# hodor
Framework for Automated DevSecOps Testing that is Scaleable and Asynchronous built on Kubernetes

![hodor](imgs/hodorarch.jpeg)

## Getting Started

### Local Deployment using Minikube (single node Kubernetes cluster) with in-cluster config

* Install docker-machine and kubectl

* git clone this repo

* Grab your Google Service Account credentials file and save it as `creds.json`. If you don't already have this file, you will need to create a service account credential (for Google App Engine) by navigating to https://console.developers.google.com/apis/credentials?project=<project-id>. Once you do that, save the .json file as creds.json.

* Install Minikube following the instructions here - http://kubernetes.io/docs/getting-started-guides/minikube/. For MacOS, it was the command `curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.12.0/minikube-darwin-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/`

* Type `minikube start` to start a single node Kubernetes cluster on your local machine

* Type `minikube dashboard` to start the Kubernetes dashboard in your browser

* Type `eval $(minikube docker-env)` to setup your docker config to point to minikube's docker environment

* Now, you can do a `docker ps -a` or `docker images` to see all the Docker containers/images of your Minikube Kubernetes cluster for troubleshooting, etc. You can always do a `docker-machine ls` and then `docker-machine env <machinename>` and then `eval` it to switch back to other docker machines, if needed.

* You can also run commands like `kubectl get nodes` or `kubectl get pods` to get started with running kubectl commands on your local Kubernetes cluster. If you want to switch `kubectl` to query a different Kubernetes cluster, you can do that by listing all the contexts by `kubectl config view` and then typing `kubectl config use-context <context-name>`

* Next, we need to build the following docker images for the different Pods that we would be starting on Kubernetes:
    * "docker build -t tools_nmap ." from the hodor/tools/nmap directory
    * "docker build -t hodor_api:v1 ." from the hodor/api directory
    * "docker build -t machinery_worker:v1 ." from the hodor/api/machinery directory
    * "docker build -t google_subscription:v1 ." from the hodor/api/googlesubscription directory

* We will be using a public RabbitMQ image so just type `docker pull rabbitmq` to make sure you have that image as well. By this point, we have built all the Docker images we need to deploy our minikube cluster.

* If you look at the `.yaml` files in the `hodor/config` directory, you will notice that I am using some environment variables for `PROJECT_ID`, `PUBSUB_TOPICNAME`, `BUCKET_NAME`, `SUBSCRIPTION_NAME`, `DATASET_NAME`, `TABLE_NAME` and `GOOGLE_APPLICATION_CREDENTIALS` in some of those files.

* Replace these values with the values corresponding to your Google Service Account. You will need your project ID on GCP, the topic and subscription name for Google PubSub, your GSC bucket name, your BigQuery dataset and table name.

* Leave the `GOOGLE_APPLICATION_CREDENTIALS` environment variable as is.

* Create a Kubernetes secret from your Google service account credentials file like this:
`kubectl create secret generic googlesecret --from-file=/path/to/creds.json`

* Make sure you can see it by typing `kubectl get secret googlesecret -o yaml`

* At this point, you should be set to start your deployments in the Kubernetes cluster.

* There are 6 yaml files right now in the `hodor/config` directory:
    * googlesubscription-deployment.yaml (deploys 2 replicas for the Google Subscription worker)
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

* Finally, send a curl request to the API service such as `curl -H "Content-Type: application/json" -X POST -d '{"Toolname":"<TOOLNAME>","Targets":["<IP1>", "<IP2>"],"Options":"<OPTIONS>"}' http://IP:PORT/api/v1/runtool`. You should receive 2 Task IDs in the response. Navigate to the worker pods and look at their logs. Notice that the workers picked up the tasks and completed it. They then sent the filename fo Google PubSub and stored the scan result file (after converting to CSV) to Google Storage Cloud. 

* The Subscription worker consumed the filename and ran another job taking the CSV data from Google Storage Cloud and uploading it to Google BigQuery. That's it!


### Local Development with out-cluster config 

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
        - export GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json
      Then, run "go run machineryworker.go". This will start the workers to pick up tasks from the RabbitMQ broker. 
    
    * In another terminal, navigate to hodor/api directory and run "go run *.go". This will start all the components of the API server. 
    
    * In another terminal, navigate to hodor/api/googlesubscription directory. You need the following environment variables for this terminal:
        - export PROJECT_ID=<PROJECTID>
        - export SUBSCRIPTION_NAME=<SUBSCRIPTIONNAME>
        - export DATASET_NAME=<DATASETNAME>
        - export TABLE_NAME=<TABLENAME>
        - export BUCKET_NAME=<BUCKETNAME>
        - export GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json
      Then, run "go run googlesubscription.go". This will start the Subscription workers. 

    * The infrastructue is up and running now. You can start sending API requests to the API server running locally on port 3636 such as "curl -H "Content-Type: application/json" -X POST -d '{"Toolname":"<TOOLNAME>","Targets":["<IP1>", "<IP2>"],"Options":"<OPTIONS>"}' http://localhost:3636/api/v1/runtool". Rest is similar as above. 


### Building your own tools in this framework

You can either contribute in building this framework or you can try integrating your own tool in this framework. 
In order to build/integrate your own tool in this framework, you would have to write a Dockerfile for it and ensure that it can be run by simply running the Docker container such as `docker run <your-tool-name>` or something like that.
Essentially, you would have to build a Docker image of your tool such that you can just send the API server a request specifying your toolname and your tool takes care of everything.
If not, you can specify the target and options as well. Some code change would have to go in to ensure that the framework works with your tool as well.


### Notes
* The entire architecture will be orchestrated by Kubernetes so the number of pods can be increased/decreased depending upon the load by simply calling the Kubernetes API with no downtime. This can also be auto scaled with future versions of Kubernetes.
* All the tools that will be run would be deployed as Docker images on the Google Container Registry.
* Right now, the RabbitMQ message broker is only 1 POD. This can be improved by deploying a clustered RabbitMQ. 
* Right now, the tools are started as single Docker containers in a POD. But, we can also start multiple Docker containers in a POD. This would depend upon what we see after doing some experimentation on what works and what doesn't.
* The tool used for the message broker and worker is called machinery - https://github.com/RichardKnop/machinery
* We should build our own Docker images and upload it to an internal registry instead of trusting public Docker images.
* We need to implement some form of authentication to the API server.
* Since we are only deploying a single node Kubernetes cluster using Minikube, I am mounting the host aka node path as the volume inside the Worker and Job Pods so that the scan files can be stored and retrieved from there. Once we start deploying clusters on GCE, we would have to replace the HostPath to something like GCE Persistent Disk.
* Right now, there is logic for only running nmap. So, if you want to integrate more tools, there will be some changes that need to go in.