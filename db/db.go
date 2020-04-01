package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type ID struct {
	BoardID      string   `json:"BoardID"`
	PageID       string   `json:"PageID"`
	FileID       string   `json:"FileID"`
	BackgroundID string   `json:"BackgroundID"`
	ImageID      []string `json:"ImageID"`
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)
	items := getItems()
	//fmt.Println(item.BoardID)
	tableName := "List"
	for _, item := range items {
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(av)
		// Create item in table
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}
func getItems() []ID {
	raw, err := ioutil.ReadFile("./../id.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var items []ID
	err = json.Unmarshal(raw, &items)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return items
}
