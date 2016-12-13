package main

import (
	"flag"
	"os"

	"../machinery/machinerytasks"
	m "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/errors"
)

var (
	configPath    = flag.String("c", "config.yml", "Path to a configuration file")
	broker        = flag.String("b", "amqp://guest:guest@rabbitmq-service:5672/", "Broker URL")
	resultBackend = flag.String("r", "amqp://guest:guest@rabbitmq-service:5672/", "Result backend")
	exchange      = flag.String("e", "machinery_exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType  = flag.String("t", "direct", "Exchange type - direct|fanout|topic|x-custom")
	defaultQueue  = flag.String("q", "machinery_tasks", "Ephemeral AMQP queue name")
	bindingKey    = flag.String("k", "machinery_task", "AMQP binding key")

	cnf    config.Config
	server *m.Server
	worker *m.Worker
)

func init() {

	flag.Parse()

	cnf = config.Config{
		Broker:        *broker,
		ResultBackend: *resultBackend,
		Exchange:      *exchange,
		ExchangeType:  *exchangeType,
		DefaultQueue:  *defaultQueue,
		BindingKey:    *bindingKey,
	}

	// Parse the config
	// NOTE: If a config file is present, it has priority over flags
	data, err := config.ReadFromFile(*configPath)
	if err == nil {
		err = config.ParseYAMLConfig(&data, &cnf)
		errors.Fail(err, "Could not parse config file")
	}

	server, err := m.NewServer(&cnf)
	errors.Fail(err, "Could not initialize server")

	// Register tasks
	tasks := map[string]interface{}{
		"RunTool": machinerytasks.RunTool,
	}
	server.RegisterTasks(tasks)

	hostname, err := os.Hostname()
	check(err)

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker = server.NewWorker(hostname)
}

func main() {
	err := worker.Launch()
	errors.Fail(err, "Could not launch worker")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
