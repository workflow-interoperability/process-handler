package worker

import (
	"context"
	"log"

	agrpc "github.com/workflow-interoperability/activity/grpc"
	"github.com/workflow-interoperability/process-handler/lib"
	"github.com/zeebe-io/zeebe/clients/go/entities"
	"github.com/zeebe-io/zeebe/clients/go/worker"
	"google.golang.org/grpc"
)

func Sync1(client worker.JobClient, job entities.Job) {
	conn, err := grpc.Dial("localhost:8082", grpc.WithInsecure())
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
	variables["sync2"] = "sync2"

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		lib.FailJob(client, job)
		return
	}

	request.Send()
}
