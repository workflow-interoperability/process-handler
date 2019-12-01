package worker

import (
	"context"
	"log"

	agrpc "github.com/workflow-interoperability/activity/grpc"
	mygrpc "github.com/workflow-interoperability/factory/grpc"
	"github.com/workflow-interoperability/process-handler/lib"
	"github.com/zeebe-io/zeebe/clients/go/entities"
	"github.com/zeebe-io/zeebe/clients/go/worker"
	"google.golang.org/grpc"
)

func StartSub(client worker.JobClient, job entities.Job) {
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}
	sRClient := mygrpc.NewServiceRegistryClient(conn)
	fClient := mygrpc.NewFactoryClient(conn)
	// get sub id
	fs, err := sRClient.ListDefinitions(context.Background(), &mygrpc.ServiceRegistryListDefinitionsRq{})
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}
	var id string
	for _, v := range fs.GetSequence() {
		if v.Name == "sub" {
			id = v.DefinitionKey
			break
		}
	}
	if id == "" {
		log.Println("can not get sub definition")
		lib.FailJob(client, job)
		return
	}
	// create sub instance
	res, err := fClient.CreateInstance(context.Background(), &mygrpc.FactoryCreateInstanceRq{
		StartImmediately: true,
		ObserverKey:      "Hello",
		Name:             "sub_1",
		Subject:          "Hello",
		Description:      "sub instance",
		ContextData:      "{}",
		FactoryKey:       id,
	})
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}

	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}
	// variables["sync1"] = "sync1"

	variables["subid"] = res.GetInstanceKey()
	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		lib.FailJob(client, job)
		return
	}

	request.Send()
}

func Sync2(client worker.JobClient, job entities.Job) {
	conn, err := grpc.Dial("localhost:8083", grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}
	// notify observer state has been changed
	aClient := agrpc.NewActivityClient(conn)
	_, err = aClient.CompleteActivity(context.Background(), &agrpc.ActivityCompleteActivityRq{
		Key: job.GetType(),
	})
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}

	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		log.Println(err)
		lib.FailJob(client, job)
		return
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		lib.FailJob(client, job)
		return
	}

	request.Send()
}
