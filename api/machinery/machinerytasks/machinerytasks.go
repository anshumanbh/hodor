package machinerytasks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"os"

	"io/ioutil"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	gnmap "github.com/lair-framework/go-nmap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	jobv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"

	"strconv"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func create(x int32) *int32 {
	return &x
}

func i(o int64) *int64 {
	return &o
}

func c(y bool) *bool {
	return &y
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	return v
}

// RunTool ...
func RunTool(toolname string, targets string, options string) (string, error) {

	config, err := rest.InClusterConfig()
	check(err)

	// config, err := clientcmd.BuildConfigFromFlags("", "/Users/abhartiya/.kube/config")
	// check(err)

	clientset, err := kubernetes.NewForConfig(config)
	check(err)

	opt := strings.Split(options, " ")
	optarray := make([]string, len(opt))
	for index, element := range opt {
		optarray[index] = element
	}

	var buffer bytes.Buffer
	var flocbuffer bytes.Buffer
	if toolname == "tools_nmap" {
		buffer.WriteString("/results/nmap_")
		buffer.WriteString(targets)
		buffer.WriteString(".xml")

		optarray = append(optarray, "-oX")
		optarray = append(optarray, buffer.String())
	}

	pwd, err := os.Getwd()
	check(err)

	runtooljob, err := clientset.Batch().Jobs(api.NamespaceDefault).Create(&jobv1.Job{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "tool-job",
		},
		Spec: jobv1.JobSpec{
			Parallelism: create(1),
			Completions: create(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app":       "taskQueue",
						"component": "jobs",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "tool-container",
							Image:           toolname,
							ImagePullPolicy: "IfNotPresent",
							Args:            append(optarray, targets),
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "results",
									MountPath: "/results",
								},
							},
						},
					},

					RestartPolicy: "OnFailure",
					Volumes: []v1.Volume{
						{
							Name: "results",
							VolumeSource: v1.VolumeSource{
								// GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
								// 	PDName: "scan-results-disk",
								// 	FSType: "ext4",
								// },
								// EmptyDir: &v1.EmptyDirVolumeSource{
								// 	Medium: "",
								// },
								// saving the results to the host for now. this can be changed. there are different volume options that can be used
								HostPath: &v1.HostPathVolumeSource{
									Path: pwd + "/results",
									// Path: "/api/results",
								},
							},
						},
					},
				},
			},
		},
	})
	check(err)

	fmt.Printf("Job submitted: %s", runtooljob.Name)
	fmt.Println("")

	for {
		cjob, err := clientset.Batch().Jobs(api.NamespaceDefault).Get(runtooljob.Name)
		check(err)

		cjobstatus := cjob.Status.Succeeded
		fmt.Printf("Job: %s Status: %d", runtooljob.Name, cjobstatus)
		fmt.Println("")

		if cjobstatus == 1 {
			break
		}
		time.Sleep(10 * time.Second)
	}

	fjob, err := clientset.Batch().Jobs(api.NamespaceDefault).Get(runtooljob.Name)
	check(err)

	cpods, err := clientset.Core().Pods(api.NamespaceDefault).List(v1.ListOptions{})
	check(err)

	ctx := context.Background()
	pubsubclient, err := pubsub.NewClient(ctx, mustGetenv("PROJECT_ID"))
	check(err)
	topicName := mustGetenv("PUBSUB_TOPICNAME")
	topic := pubsubclient.Topic(topicName)

	sclient, err := storage.NewClient(ctx)
	bucket := sclient.Bucket(mustGetenv("BUCKET_NAME"))

	defer func() {

		if toolname == "tools_nmap" {

			fmt.Println("Ran NMAP and saved the XML file")
			flocbuffer.WriteString(pwd)
			flocbuffer.WriteString(buffer.String())

			xmlfile, err := ioutil.ReadFile(flocbuffer.String())
			check(err)

			fmt.Println("Parsing the XML and converting into CSV for BQ")

			nmaprun, err := gnmap.Parse(xmlfile)
			check(err)

			//creating a csv filename
			var filename bytes.Buffer
			filename.WriteString("nmap_")
			filename.WriteString(targets)
			filename.WriteString(".csv")
			csvfilename := filename.String()

			wc := bucket.Object(csvfilename).NewWriter(ctx)

			var t bytes.Buffer

			//writing the data to a csv file directly on GCS
			for _, host := range nmaprun.Hosts {
				for _, ip := range host.Addresses {
					for _, port := range host.Ports {

						t.WriteString(ip.Addr)
						t.WriteString(",")
						t.WriteString(strconv.Itoa(port.PortId))
						t.WriteString(",")
						t.WriteString(port.Protocol)
						t.WriteString(",")
						t.WriteString(port.Service.Name)
						t.WriteString(",")
						t.WriteString(port.State.State)
						t.WriteString("\n")
						barray := []byte(t.String())
						_, err := wc.Write(barray)
						check(err)
					}
				}
				defer wc.Close()
			}

			fmt.Println("CSV file stored in the bucket as an Object")

			//publishing the csvfile to the google pubsub topic
			fmt.Println("Publishing the object name to the Google PubSub topic")
			msgIDs, err := topic.Publish(ctx, &pubsub.Message{
				Data: []byte(csvfilename),
			})
			check(err)

			for _, id := range msgIDs {
				fmt.Printf("Published the message: %s ", csvfilename)
				fmt.Printf("msg ID: %v\n", id)
			}

			//TODO: delete the xml file

		}

		if toolname != "tools_nmap" {
			fmt.Println("Ran a different tool. Need logic to send the result file from this tool")
		}

		// Delete Job that finished running
		if fjob.Status.Succeeded == 1 {
			fmt.Printf("Job: %s Status: %d", runtooljob.Name, fjob.Status.Succeeded)
			fmt.Println("")
			fmt.Printf("Deleting Job: %s", runtooljob.Name)
			fmt.Println("")
			clientset.Batch().Jobs(api.NamespaceDefault).Delete(runtooljob.Name, &v1.DeleteOptions{})

			//Delete PODs that have been completed where the jobs ran
			for _, element := range cpods.Items {
				if element.Status.Phase == "Succeeded" && element.Labels["app"] == "taskQueue" && element.Labels["component"] == "jobs" {
					fmt.Printf("Job Pod: %s Status: %s", element.Name, element.Status.Phase)
					fmt.Println("")
					fmt.Printf("Deleting Job Pod: %s", element.Name)
					clientset.Core().Pods(api.NamespaceDefault).Delete(element.Name, &v1.DeleteOptions{})
					fmt.Println("")
				}
			}

		}
	}()

	return "Job Successful", nil

}
