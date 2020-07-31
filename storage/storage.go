package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

var Session *session.Session
var S3Service *s3.S3
var Buckets []*s3.Bucket

func init() {
	Session = session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("https://us-east-1.linodeobjects.com"),
		Region:   aws.String("us-east-1"),
	}))
	S3Service = s3.New(Session)

	log.Println("EC2 client initialized... testing...")

	tempBuckets, err := S3Service.ListBuckets(nil)
	if err != nil {
		log.Fatal("Unable to list buckets ", err.Error())
	}

	Buckets = tempBuckets.Buckets

	log.Println("EC2 test passed")
}
