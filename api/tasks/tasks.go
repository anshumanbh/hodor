package tasks

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	jobv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func create(x int32) *int32 {
	return &x
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	return v
}

// TestTask ...
func TestTask() (string, error) {
	return "test", nil
}

// RunTool ...
func RunTool(toolname string, targets string, options string) (string, error) {

	config, err := rest.InClusterConfig()
	check(err)

	// config, err := clientcmd.BuildConfigFromFlags("", "/path/to/.kube/config")
	// check(err)

	clientset, err := kubernetes.NewForConfig(config)
	check(err)

	opt := strings.Split(options, " ")
	optarray := make([]string, len(opt))
	for index, element := range opt {
		optarray[index] = element
	}

	if toolname == "tools_nmap" {

		var buffer bytes.Buffer
		buffer.WriteString("/results/nmap_")
		buffer.WriteString(targets)
		buffer.WriteString(".xml")

		optarray = append(optarray, "-oX")
		optarray = append(optarray, buffer.String())
	}

	fmt.Println(optarray)

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
						"app":       "taskqueue",
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
								// saving the results to the host for now. this can be changed. there are different volume options that can be used
								HostPath: &v1.HostPathVolumeSource{
									//this creates a results directory inside /hodor/api/worker directory
									Path: pwd + "/results",
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

	defer func() {

		// Delete Job that finished running
		if fjob.Status.Succeeded == 1 {
			fmt.Printf("Job: %s Status: %d", runtooljob.Name, fjob.Status.Succeeded)
			fmt.Println("")
			fmt.Printf("Deleting Job: %s", runtooljob.Name)
			fmt.Println("")
			clientset.Batch().Jobs(api.NamespaceDefault).Delete(runtooljob.Name, &v1.DeleteOptions{})

			//Delete PODs that have been completed where the jobs ran
			for _, element := range cpods.Items {
				if element.Status.Phase == "Succeeded" && element.Labels["app"] == "taskqueue" && element.Labels["component"] == "jobs" {
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
