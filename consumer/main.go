package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	svc := sqs.New(session.New(), &aws.Config{
		Endpoint: aws.String(os.Getenv("SQS_SERVER")),
		Region:   aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			"id",
			"secret",
			"token",
		)})
	r.Use(middleware.Logger)
	go func() {
		for {
			queueURL := os.Getenv("SQS_SERVER") + "/queue/" + os.Getenv("QUEUE_URL")
			result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
				AttributeNames: []*string{
					aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
				},
				MessageAttributeNames: []*string{
					aws.String(sqs.QueueAttributeNameAll),
				},
				QueueUrl:          aws.String(queueURL),
				VisibilityTimeout: aws.Int64(20),
				WaitTimeSeconds:   aws.Int64(10),
			})
			if err != nil {
				log.Println("Fail to receive message: ", err)
			} else if len(result.Messages) > 0 {
				for _, m := range result.Messages {
					log.Println("Message received: ", m)

					_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      aws.String(queueURL),
						ReceiptHandle: m.ReceiptHandle,
					})

					if err != nil {
						log.Println(err.Error())
					}
				}
			}
			time.Sleep(time.Second)
		}
	}()

	r.Get("/send", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("consumer working"))
	})
	log.Println("Consumer running...")
	http.ListenAndServe(":3001", r)
}
