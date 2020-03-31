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
	bucket := "vibe.dev"
	/*
		testitem := "boards/Ct7dNTKpmr26rGb0PrW1L0/7Wufrt1lF52lfoH6WGQ1vC/page.vpf"
		buf, _ := downloadFile(sess, bucket, testitem)
		page := &pb.PageFile{}
		if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
			fmt.Println(err)
		}
		listImageID(buf, testitem)
	*/
	poll(sess, bucket)
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

		for i, item := range resp.Contents {
			if strings.HasSuffix(*item.Key, "vpf") {
				buf, err := downloadFile(sess, bucket, *item.Key)
				if err != nil {
					return err
				}
				if err = listImageID(buf, *item.Key); err != nil {
					return err
				}
				fmt.Printf("%d\n", i)
			}
		}
		if *resp.NextContinuationToken != "" {
			params.ContinuationToken = resp.NextContinuationToken
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

func listImageID(buf *aws.WriteAtBuffer, item string) error {
	page := &pb.PageFile{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		return err
	}

	for _, event := range page.GetPageEvents() {
		flag := false
		if img := event.GetAddMagnet().GetMagnet().GetImage(); img != nil {
			mgtImgID := img.GetImageId()
			flag = true
			defer fmt.Println("magnetImageID:", mgtImgID)
		}

		if bkgImg := event.GetSetBackground(); bkgImg != nil {
			bkgImgID := bkgImg.GetImageId()
			flag = true
			defer fmt.Println("backgroundImageID:", bkgImgID)
		}

		if file := event.GetSetBackground().GetBackgroundMeta().GetPdfMeta(); file != nil {
			fileID := file.GetFileId()
			flag = true
			defer fmt.Println("fileID:", fileID)
		}
		if flag {
			fmt.Println("boardID:", strings.FieldsFunc(item, f)[1])
			fmt.Println("PageID:", page.GetPageId())
		}
	}

	return nil
}
func f(c rune) bool {
	return c == '/'
}

/*
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
*/
