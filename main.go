package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"vibe/pb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/protobuf/proto"
)

//ID
type ID struct {
	BoardID string   `json:"BoardID"`
	PageID  string   `json:"PageID"`
	FileID  []string `json:"FileID"`
}

func main() {
	accessKey := ""
	secretKey := ""

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(endpoints.ApSoutheast1RegionID),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))
	bucket := "vibe.dev"
	/*
		testitem := "boards/Ct7dNTKpmr26rGb0PrW1L0/7Wufrt1lF52lfoH6WGQ1vC/page.vpf"
		buf, _ := downloadFile(sess, bucket, testitem)
		page := &pb.PageFile{}
		if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
			fmt.Println(err)
		}
		listFileID(buf, testitem)
	*/
	poll(sess, bucket)
}

func poll(sess *session.Session, bucket string) error {
	svc := s3.New(sess)

	//	fmt.Printf(bucket)
	//	fmt.Printf("\n")

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
			if strings.HasSuffix(*item.Key, "vpf") {
				buf, err := downloadFile(sess, bucket, *item.Key)
				if err != nil {
					return err
				}
				if err = listFileID(buf, *item.Key); err != nil {
					return err
				}
				//fmt.Printf("\n")
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

	_, err := downloader.Download(buf, input)
	if err != nil {
		return nil, err
	}
	//fmt.Println("file downloaded", numBytes, "bytes")*/
	return buf, nil
}

func listFileID(buf *aws.WriteAtBuffer, item string) error {
	page := &pb.PageFile{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		return err
	}
	flag := false
	id := ID{}
	id.BoardID = strings.Split(item, "/")[1]
	id.PageID = strings.Split(item, "/")[2]
	for _, event := range page.GetPageEvents() {
		if img := event.GetAddMagnet().GetMagnet().GetImage(); img != nil {
			id.FileID = append(id.FileID, img.GetImageId())
			flag = true
		}
		if bkgImg := event.GetSetBackground(); bkgImg != nil {
			id.FileID = append(id.FileID, bkgImg.GetImageId())
			flag = true
		}
		if file := event.GetSetBackground().GetBackgroundMeta().GetPdfMeta(); file != nil {
			id.FileID = append(id.FileID, file.GetFileId())
			flag = true
		}
	}
	if flag {
		idJSON, _ := json.Marshal(id)
		fmt.Println(string(idJSON))
		return nil
	}
	return nil
}

/*
func listBackground(buf *aws.WriteAtBuffer) error {
	page := &pb.Background{}
	if err := proto.Unmarshal(buf.Bytes(), page); err != nil {
		return err
	}

	if bkgimg := page.GetImageId(); bkgimg != "" {
		fmt.Println("backgroundFileID:", bkgimg)
	} else {
		fmt.Println("backgroundFileID:No backgroundFileID")
	}

	if fileID := page.GetMeta().GetPdfMeta().GetFileId(); fileID != "" {
		fmt.Println("fileid:", fileID)
	} else {
		fmt.Println("fileid:No fileID")
	}
	return nil
}
*/
