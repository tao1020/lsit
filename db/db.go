package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type ID struct {
	BoardID string   `json:"BoardID"`
	PageID  string   `json:"PageID"`
	FileID  []string `json:"FileID"`
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)
	tableName := "List"
	file, error := os.OpenFile("./../id.json", os.O_RDONLY, 0)
	if error != nil {
		panic(error)
	}
	defer file.Close()

	raw := bufio.NewScanner(file)
	for raw.Scan() {
		var item ID
		err := json.Unmarshal(raw.Bytes(), &item)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Create item in table Movies
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
