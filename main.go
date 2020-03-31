package main

import (
	"fmt"
	"strings"
	"vibe/pb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/protobuf/proto"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.ApSoutheast1RegionID),
	}))

	if err := poll(sess, "vibe.dev"); err != nil {
		fmt.Println("fail to pool files", err)
	}
}

func poll(sess *session.Session, bucket string) error {
	svc := s3.New(sess)

	fmt.Printf(bucket)
	fmt.Printf("\n")

	params := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucket),
		Prefix:            aws.String("boards/"),
		ContinuationToken: nil,
	}
	for {
		resp, err := svc.ListObjectsV2(params)
		if err != nil {
			return err
		}

		for _, item := range resp.Contents {
			if strings.Contains(*item.Key, ".vpf") {
				buf, err := downloadFile(sess, bucket, *item.Key)
				if err != nil {
					return err
				}
				if err = listBackground(buf); err != nil {
					return err
				}
				if err = listImageID(buf); err != nil {
					return err
				}
				fmt.Printf("\n")
			}
		}
		if resp.ContinuationToken != nil {
			params.ContinuationToken = resp.ContinuationToken
		} else {
			break
		}
	}
	return nil
}

func downloadFile(sess *session.Session, bucket string, item string) (*aws.WriteAtBuffer, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	}
	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	numBytes, err := downloader.Download(buf, input)
	if err != nil {
		return nil, err
	}
	fmt.Println("file downloaded", numBytes, "bytes")
	return buf, nil
}

func listBackground(buf *aws.WriteAtBuffer) error {
	page := &pb.Background{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		return err
	}

	if bkgimg := page.GetImageId(); bkgimg != "" {
		fmt.Println("backgroundImageID:", bkgimg)
	} else {
		fmt.Println("backgroundImageID:No backgroundImageID")
	}

	if fileID := page.GetMeta().GetPdfMeta().GetFileId(); fileID != "" {
		fmt.Println("fileid:", fileID)
	} else {
		fmt.Println("fileid:No fileID")
	}
	return nil
}

func listImageID(buf *aws.WriteAtBuffer) error {
	page := &pb.PageFile{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		return err
	}
	flag := true
	for _, event := range page.GetPageEvents() {
		if img := event.GetAddMagnet().GetMagnet().GetImage(); img != nil {
			fmt.Println("imageid:", img.GetImageId())
			flag = false
		}
	}
	if flag {
		fmt.Println("imageid:No image")
	}
	return nil
}
