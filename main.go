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

	poll(sess, "vibe.dev")

}

func poll(sess *session.Session, bucket string) {
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
			fmt.Printf("Unable to list items in bucket %q, %v", bucket, err)
		}

		for _, item := range resp.Contents {
			if strings.Contains(*item.Key, ".vpf") {
				buf := downloadFile(sess, bucket, *item.Key)
				listBackground(buf)
				listImageID(buf)
				fmt.Printf("\n")
			}
		}
		if resp.ContinuationToken != nil {
			params.ContinuationToken = resp.ContinuationToken
		} else {
			break
		}
	}
}

func downloadFile(sess *session.Session, bucket string, item string) *aws.WriteAtBuffer {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	}
	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	numBytes, err := downloader.Download(buf, input)
	if err != nil {
		fmt.Println("Unable to download item ", item, err)
	}
	fmt.Println("file downloaded", numBytes, "bytes")
	return buf
}

func listBackground(buf *aws.WriteAtBuffer) {
	page := &pb.Background{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		fmt.Println("failed to parse file:", err)
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
}

func listImageID(buf *aws.WriteAtBuffer) {
	page := &pb.PageFile{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		fmt.Println("failed to parse file:", err)
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
}
