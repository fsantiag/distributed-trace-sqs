package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func setupGlobalTracer() {
	// zipkin / opentracing specific stuff
	// set up a span reporter
	var reporterOpts []zipkinhttp.ReporterOption
	reporterOpts = append(reporterOpts, zipkinhttp.Logger(log.New(ioutil.Discard, "", log.LstdFlags)))
	reporter := zipkinhttp.NewReporter("http://zipkin:9411/api/v2/spans", reporterOpts...)

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint("producer", "localhost:3000")
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	log.Println(tracer)
	opentracing.SetGlobalTracer(tracer)
}

type CustomMessageInput sqs.SendMessageInput

func (c CustomMessageInput) Set(key, val string) {
	if c.MessageAttributes == nil {
		c.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
	}
	c.MessageAttributes[key] = &sqs.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(val)}
	log.Println(c.MessageAttributes)
}

func main() {
	setupGlobalTracer()
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
		span := opentracing.StartSpan("Producer: first span")
		defer span.Finish()
		queueUrl := os.Getenv("SQS_SERVER") + "/queue/" + os.Getenv("QUEUE_URL")
		params := CustomMessageInput{
			MessageBody: aws.String("Testing 1,2,3,..."), // Required
			QueueUrl:    aws.String(queueUrl),            // Required
		}
		err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapWriter(params))
		if err != nil {
			log.Fatal(err)
		}
		_, err = svc.SendMessage((*sqs.SendMessageInput)(&params))
		if err != nil {
			log.Println(err)
		}
		w.Write([]byte("sent to sqs"))
	})
	log.Println("Producer running...")
	http.ListenAndServe(":3000", r)
}
