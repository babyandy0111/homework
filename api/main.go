package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"context"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type SQSSendMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type MyFriend struct {
	Name   string `json:"name"`
	UserID int    `json:"id"`
}

type User struct {
	Name       string `json:"name"`
	UserID     int    `json:"user_id"`
	ActionType string `json:"action_type"`
}

type SendData struct {
	FromID     int    `json:"from_id"`
	ToID       int    `json:"to_id"`
	ActionType string `json:"action_type"`
}

var db *sql.DB

func main() {
	r := gin.Default()

	// 使用request cache 當第一道cache機制，可選放在電腦或redis
	store := persistence.NewInMemoryStore(time.Second)

	// 寫route，預期應該會有以下幾隻API
	// 取得該user所有好友
	r.GET("/:user_id/friend", func(c *gin.Context) {
		Init()
		userID, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error user_id": err,
			})
			return
		}

		qry := `select 
					user.id, user.name 
				from friend 
				left join user on user.id = friend.to_id
				where friend.from_id = ? `

		rows, err := db.Query(qry, userID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"db error": err.Error(),
			})
		} else {
			myFriend := []MyFriend{}
			for rows.Next() {
				var tmp MyFriend
				if err := rows.Scan(&tmp.UserID, &tmp.Name); err != nil {
					log.Fatal(err)
				}
				myFriend = append(myFriend, tmp)
			}
			if err := rows.Err(); err != nil {
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, gin.H{
				"data": myFriend,
			})
		}

		defer rows.Close()
		defer db.Close()
	})

	// 新增刪除user
	r.POST("/user", cache.CachePage(store, time.Second*3, userAction))
	r.DELETE("/user/:user_id", cache.CachePage(store, time.Second*3, userAction))

	// 新增刪除訂閱、新增刪除好友
	r.POST("/:user_id/:action/:friend_id", cache.CachePage(store, time.Second*3, friendOrSubscriptionAction))
	r.DELETE("/:user_id/:action/:friend_id", cache.CachePage(store, time.Second*3, friendOrSubscriptionAction))

	r.Run() // listen and serve on 0.0.0.0:8080
}

func friendOrSubscriptionAction(c *gin.Context) {
	fromID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error user_id": err,
		})
		return
	}

	toID, err := strconv.Atoi(c.Param("friend_id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error friend_id": err,
		})
		return
	}

	if c.Param("action") == "friend" || c.Param("action") == "subscription" {
		var sendData SendData
		sendData.FromID = fromID
		sendData.ToID = toID
		if c.Request.Method == "POST" {
			sendData.ActionType = fmt.Sprintf("add_%s", c.Param("action"))
		} else {
			sendData.ActionType = fmt.Sprintf("del_%s", c.Param("action"))
		}

		err = sendMessage(c.Param("action"), sendData)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"error action": err,
		})
	}
}
func userAction(c *gin.Context) {
	var user User
	err := c.Bind(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error json": err,
		})
		return
	}

	if c.Request.Method == "DELETE" {
		UserID, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error Atoi": err,
			})
			return
		}
		user.UserID = UserID
	}

	user.ActionType = c.Request.Method

	err = sendMessage("user", user)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error send": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
func sendMessage(queueName string, target interface{}) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)

	gQInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}

	result, err := GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:")
		fmt.Println(err)
		return nil
	}

	queueURL := result.QueueUrl
	jsonString, err := json.Marshal(target)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	sMInput := &sqs.SendMessageInput{
		DelaySeconds: 10,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String("The HomeWork"),
			},
			"Author": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Andy"),
			},
			"WeeksOn": {
				DataType:    aws.String("Number"),
				StringValue: aws.String("6"),
			},
		},
		MessageBody: aws.String(string(jsonString)),
		QueueUrl:    queueURL,
	}

	resp, err := SendMsg(context.TODO(), client, sMInput)
	if err != nil {
		fmt.Println("Got an error sending the message:")
		fmt.Println(err)
		return nil
	}

	fmt.Println("Sent message with ID: " + *resp.MessageId)
	return nil
}
func SendMsg(c context.Context, api SQSSendMessageAPI, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessage(c, input)
}
func GetQueueURL(c context.Context, api SQSSendMessageAPI, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return api.GetQueueUrl(c, input)
}
func Init() {
	var err error
	err = godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPW := os.Getenv("DB_PW")
	dbUser := os.Getenv("DB_USER")
	connString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPW, dbHost, dbName)
	fmt.Println(connString)
	db, err = sql.Open("mysql", connString)
	if err != nil {
		log.Fatalln(err)
	}
}
