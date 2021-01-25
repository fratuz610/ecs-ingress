package util

import (
	"bitbucket.org/nnnco/rev-proxy/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Client simplified client to access S3 resources on AWS
type S3Client struct {
	cfg          *shared.Config
	session      *session.Session
	s3downloader *s3manager.Downloader
}

// NewS3Client Creates a new s3 client
func NewS3Client(cfg *shared.Config) *S3Client {

	mySession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(cfg.AWS.Region)},
	}))

	// Create a S3 clients from just a session

	ret := &S3Client{
		cfg:          cfg,
		session:      mySession,
		s3downloader: s3manager.NewDownloader(mySession),
	}

	return ret
}

// DownloadFileInMemory exported
func (s *S3Client) DownloadFileInMemory(bucket string, key string) ([]byte, error) {

	buf := &aws.WriteAtBuffer{}

	_, err := s.s3downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
