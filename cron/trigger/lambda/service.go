package lambda

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"smirror/cron/config"
	"smirror/cron/trigger"
	"time"
)

type service struct {
	*lambda.Lambda
}

//Trigger triggers lambda execution
func (s *service) Trigger(ctx context.Context, resource *config.Resource, eventSource storage.Object) error {
	URL := eventSource.URL()
	bucket := url.Host(URL)
	URLPath := url.Path(URL)
	s3Event := events.S3Event{Records: make([]events.S3EventRecord, 0)}
	s3Event.Records = append(s3Event.Records, events.S3EventRecord{
		AWSRegion:   resource.Region,
		EventTime:   time.Now(),
		EventSource: "s3",
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: bucket,
			},
			Object: events.S3Object{
				Key:URLPath,
				Size:eventSource.Size(),
			},
		},
	})
	payload, err := json.Marshal(s3Event)
	if err != nil {
		return errors.Wrapf(err, "failed to decode s3 event for %v", eventSource.URL())
	}
	_, err = s.Invoke(&lambda.InvokeInput{
		FunctionName:&resource.DestFunction,
		Payload:payload,
		InvocationType:aws.String(lambda.InvocationTypeEvent),
	})
	return err

}

//New create trigger service
func New() (trigger.Service, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &service{Lambda: lambda.New(sess)}, nil
}
