package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

var db *sql.DB

func init() {
	var err error
	dbHost := os.Getenv("db_host")
	dbName := os.Getenv("db_name")
	dbPW := os.Getenv("db_pw")
	dbUser := os.Getenv("db_user")
	connString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPW, dbHost, dbName)

	db, err = sql.Open("mysql", connString)
	if err != nil {
		log.Fatalln(err)
	}

	// 不要關閉，這樣才能重複使用
	// defer db.Close()
}

func main() {
	lambda.Start(HandleRequest)
}

type Friend struct {
	FromID     int       `json:"from_id"`
	ToID       int       `json:"to_id"`
	ActionType string    `json:"action_type"`
	CreatedDT  time.Time `json:"created_dt"`
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) (events.APIGatewayProxyResponse, error) {
	for _, message := range sqsEvent.Records {
		fmt.Printf("The message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)

		var friend Friend
		json.Unmarshal([]byte(message.Body), &friend)

		lc, _ := lambdacontext.FromContext(ctx)
		fmt.Println("reqID:", lc.AwsRequestID)

		if friend.ActionType == "add_friend" {
			qry := `INSERT INTO friend(from_id, to_id, created_at) VALUES (?, ?, ?)`
			if _, err := db.ExecContext(ctx, qry, friend.FromID, friend.ToID, time.Now().UTC().Format("2006-01-02 03:04:05")); err != nil {
				fmt.Println("err:", err.Error())
				return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
			}
		}

		if friend.ActionType == "del_friend" {
			qry := `DELETE FROM friend WHERE from_id = ? and to_id = ? `
			if _, err := db.ExecContext(ctx, qry, friend.FromID, friend.ToID); err != nil {
				fmt.Println("err:", err.Error())
				return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
			}
		}

		log := `INSERT INTO log(user_id, action, to_id, created_at) VALUES (?, ?, ?, ?)`
		if _, err := db.ExecContext(ctx, log, friend.FromID, friend.ActionType, friend.ToID, time.Now().UTC().Format("2006-01-02 03:04:05")); err != nil {
			fmt.Println("err:", err.Error())
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "ok",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 401,
		Body:       "can not request from url",
	}, nil
}
