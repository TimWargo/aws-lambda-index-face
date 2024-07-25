package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

type Event struct {
	S3Bucket string `json:"s3Bucket"`
	S3Key    string `json:"s3Key"`
}

type Response struct {
	BoundingBox BoundingBox `json:"BoundingBox"`
	Confidence  float64     `json:"Confidence"`
	FaceId      string      `json:"FaceId"`
	ImageId     string      `json:"ImageId"`
}

type BoundingBox struct {
	Height float64 `json:"Height"`
	Left   float64 `json:"Left"`
	Top    float64 `json:"Top"`
	Width  float64 `json:"Width"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event Event) (Response, error) {
	// Generate aws session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	// Retrieve process variables
	srcKey, err := url.QueryUnescape(event.S3Key)
	if err != nil {
		return Response{}, fmt.Errorf("failed to decode S3 key: %v", err)
	}
	collectionId := os.Getenv("REKOGNITION_COLLECTION_ID")

	// Create rekognition client
	rekognitionClient := rekognition.New(sess)

	// Create input for index faces command
	indexFacesCommand := rekognition.IndexFacesInput{
		CollectionId: &collectionId,
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: &event.S3Bucket,
				Name:   &srcKey,
			},
		},
	}

	// Call index faces command and check for error
	response, err := rekognitionClient.IndexFacesWithContext(ctx, &indexFacesCommand)
	if err != nil {
		return Response{}, fmt.Errorf("failed to index new face: %v", err)
	}

	// Return response containing indexed face
	return Response{
		BoundingBox: BoundingBox{
			Height: *response.FaceRecords[0].Face.BoundingBox.Height,
			Left:   *response.FaceRecords[0].Face.BoundingBox.Left,
			Top:    *response.FaceRecords[0].Face.BoundingBox.Top,
			Width:  *response.FaceRecords[0].Face.BoundingBox.Width,
		},
		Confidence: *response.FaceRecords[0].Face.Confidence,
		FaceId:     *response.FaceRecords[0].Face.FaceId,
		ImageId:    *response.FaceRecords[0].Face.ImageId,
	}, nil
}
