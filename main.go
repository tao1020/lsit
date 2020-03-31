package main

import (
	"fmt"
	"io/ioutil"
	"os"
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

	listFile(sess, "vibe.dev/")

}

func listFile(sess *session.Session, bucket string) {
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
			if str := *item.Key; str[len(str)-4:] == ".vpf" {
				downloadFile(sess, bucket, str)
				page := parse()
				listID(page)
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

func downloadFile(sess *session.Session, bucket string, item string) {
	file, err := os.Create("page.vpf.buf")
	defer file.Close()
	if err != nil {
		fmt.Println("Unable to open file :", err)
	}
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		fmt.Println("Unable to download item ", item, err)
	}
	fmt.Println("downloaded", file.Name(), numBytes, "bytes")
}

func parse() (page *pb.Page) {
	file, err := ioutil.ReadFile("page.vpf.buf")
	if err != nil {
		fmt.Println("faild to read file:", err)
	}
	page = &pb.Page{}
	if err := proto.Unmarshal(file, page); err != nil {
		fmt.Println("failed to parse file:", err)
	}
	return page
}
func listID(page *pb.Page) {
	file, err := ioutil.ReadFile("page.vpf.buf")
	if err != nil {
		fmt.Println("faild to read file:", err)
	}
	page = &pb.Page{}
	if err := proto.Unmarshal(file, page); err != nil {
		fmt.Println("failed to parse file:", err)
	}
	fmt.Println("PageID:", page.GetPageId())
	if fileID := page.GetBackground().GetMeta().GetPdfMeta().GetFileId(); fileID != "" {
		fmt.Println("FileId:", fileID)
	} else {
		fmt.Println("no background file")
	}

	if bkgImage := page.GetBackground().GetImageId(); bkgImage != "" {
		fmt.Println("background imageId", bkgImage)
	} else {
		fmt.Println("no background image")
	}

	magnets := page.GetMagnets()
	count := 1
	for _, a := range magnets {
		if magnetImageID := a.GetImage().GetImageId(); magnetImageID != "" {
			fmt.Println("magnetImage", count, "ID:", magnetImageID)
			count++
		}
	}
	if count == 1 {
		fmt.Println("no magnet image")
	}
}
