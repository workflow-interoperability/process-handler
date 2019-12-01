package main

import (
	"log"

	"github.com/workflow-interoperability/process-handler/worker"
	"github.com/zeebe-io/zeebe/clients/go/zbc"
)

const brokerAddr = "0.0.0.0:26500"

func main() {
	client, err := zbc.NewZBClientWithConfig(&zbc.ZBClientConfig{
		GatewayAddress:         brokerAddr,
		UsePlaintextConnection: true,
	})

	if err != nil {
		panic(err)
	}

	stopChan := make(chan bool, 1)
	go func() {
		jobWorker := client.NewJobWorker().JobType("sync1").Handler(worker.Sync1).Open()
		defer jobWorker.Close()
		jobWorker.AwaitClose()
		log.Println("sync1 finished")
	}()

	go func() {
		jobWorker := client.NewJobWorker().JobType("startSub").Handler(worker.StartSub).Open()
		defer jobWorker.Close()
		jobWorker.AwaitClose()
		log.Println("startSub finished")
	}()

	go func() {
		jobWorker := client.NewJobWorker().JobType("sync2").Handler(worker.Sync2).Open()
		defer jobWorker.Close()
		jobWorker.AwaitClose()
		log.Println("sync2 finished")
	}()

	<-stopChan
}
