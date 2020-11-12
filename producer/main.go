package main

import (
	"log"
	"net/http"
	"os"

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
	r.Get("/send", func(w http.ResponseWriter, r *http.Request) {
		queueUrl := os.Getenv("SQS_SERVER") + "/queue/" + os.Getenv("QUEUE_URL")
		params := &sqs.SendMessageInput{
			MessageBody: aws.String("Testing 1,2,3,..."), // Required
			QueueUrl:    aws.String(queueUrl),            // Required
		}
		_, err := svc.SendMessage(params)
		if err != nil {
			log.Println(err)
		}
		w.Write([]byte("sent to sqs"))
	})
	log.Println("Producer running...")
	http.ListenAndServe(":3000", r)
}
