package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/errors"
	"github.com/RichardKnop/machinery/v1/signatures"
)

var (
	server *machinery.Server

	cnf = config.Config{
		Broker:        "amqp://guest:guest@rabbitmq-service:5672/",
		ResultBackend: "amqp://guest:guest@rabbitmq-service:5672/",
		Exchange:      "machinery_exchange",
		ExchangeType:  "direct",
		DefaultQueue:  "machinery_tasks",
		BindingKey:    "machinery_task",
	}

	// path = "/tmp/iplist.txt"
)

type Hodor struct {
	Toolname string   `json:"toolname"`
	Targets  []string `json:"targets"`
	Options  string   `json:"options"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func RunTool(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	check(err)

	var hodor Hodor
	err = json.Unmarshal(body, &hodor)
	check(err)

	Toolname := hodor.Toolname
	Targets := hodor.Targets
	Options := hodor.Options

	server, err := machinery.NewServer(&cnf)
	errors.Fail(err, "Could not initialize server")

	tasks := make([]*signatures.TaskSignature, len(Targets))

	for i, element := range Targets {

		tasks[i] = &signatures.TaskSignature{
			Name: "RunTool",
			Args: []signatures.TaskArg{
				{
					Type:  "string",
					Value: Toolname,
				},
				{
					Type:  "string",
					Value: element,
				},
				{
					Type:  "string",
					Value: Options,
				},
			},
		}

	}

	group := machinery.NewGroup(tasks...)
	asyncResults, err := server.SendGroup(group)
	check(err)

	// taskstate := asyncResult.GetState()
	for _, asyncResult := range asyncResults {
		// result, err := asyncResult.Get()
		// check(err)
		// fmt.Println(result.Interface())
		resultstate := asyncResult.GetState()
		w.Write([]byte("RunTool job was submitted. The task ID is " + resultstate.TaskUUID + "\n"))
	}

}
