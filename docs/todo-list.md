### ToDo List / Notes:

* The entire architecture will be orchestrated by Kubernetes so the number of pods can be increased/decreased depending upon the load by simply calling the Kubernetes API with no downtime. This can also be auto scaled with future versions of Kubernetes.

* All the tools that will be run would be deployed as Docker images on the Google Container Registry.

* Right now, the RabbitMQ message broker is only 1 POD. This can be improved by deploying a clustered RabbitMQ. 

* Right now, the tools are started as single Docker containers in a POD. But, we can also start multiple Docker containers in a POD. This would depend upon what we see after doing some experimentation on what works and what doesn't.

* The tool used for the message broker and worker is called machinery - https://github.com/RichardKnop/machinery

* We should build our own Docker images and upload it to an internal registry instead of trusting public Docker images.

* We need to implement some form of authentication to the API server.

* Since we are only deploying a single node Kubernetes cluster using Minikube, I am mounting the host aka node path as the volume inside the Worker and Job Pods so that the scan files can be stored and retrieved from there. Once we start deploying clusters on GCE, we would have to replace the HostPath to something like GCE Persistent Disk.

* Right now, there is logic for only running nmap. So, if you want to integrate more tools, there will be some changes that need to go in.