package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/gokitcloud/ginkit"
	"github.com/ipfs-scrape/worker/backend"
	"github.com/ipfs-scrape/worker/ipfs"
	"github.com/ipfs-scrape/worker/queue"
	"github.com/sirupsen/logrus"
)

func main() {
	dynamodbName, ok := os.LookupEnv("IPFS_DYNAMODB_NAME")
	if !ok {
		logrus.Fatal("IPFS_DYNAMODB_NAME environment variable not set")
	}

	// Use the IPFS_DYNAMODB_NAME environment variable
	logrus.Infof("IPFS_DYNAMODB_NAME: %s", dynamodbName)

	sess, err := session.NewSession()
	if err != nil {
		logrus.Fatal(err)
	}

	svc := dynamodb.New(sess)
	dynamodbBackend, err := backend.NewDynamoDBBackend(dynamodbName, svc)
	if err != nil {
		logrus.Fatal(err)
	}

	dynamodbQueue, err := queue.NewDynamoDBQueue(dynamodbName, "ipfs", svc)
	if err != nil {
		logrus.Fatal(err)
	}

	r := ginkit.Default()
	r.GET("/ping", "pong")
	r.GET("/tokens", func(p ginkit.Params) (any, error) {
		items, err := dynamodbBackend.Scan("d-")
		logrus.Info("retrieved full dataset")
		return items, err
	})
	r.GET("/tokens/:cid", func(p ginkit.Params) (any, error) {
		if cid, ok := p.Get("cid"); ok {
			logrus.WithField("cid", cid).Info("cid")
			item, err := dynamodbBackend.Read(ipfs.GenerateIDFromCID(cid))
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			return item, nil
		}
		return nil, errors.New("invalid")
	})

	r.POST("/tokens/:cid", func(p ginkit.Params) (any, error) {
		if cid, ok := p.Get("cid"); ok {
			err := dynamodbQueue.AddItem(queue.NewQueueItem(cid, map[string]any{
				"cids": []string{cid},
			}))
			return "OK", err
		}
		return nil, errors.New("invalid")
	})

	r.POST("/bulk", func(c *gin.Context) {
		var data struct {
			Cids []string `json:"cids"`
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if len(data.Cids) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing cids"})
			return
		}

		err := dynamodbQueue.AddItem(queue.NewQueueItem(fmt.Sprintf("bulk-%v", time.Now().UnixNano()), map[string]interface{}{
			"cids": data.Cids,
		}))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to queue"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "OK"})

	})
	r.POST("/csv", func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing file"})
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = 1

		var cids []string
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid CSV file"})
				return
			}
			cids = append(cids, record[0])

			if len(cids) >= 5 {
				err = dynamodbQueue.AddItem(queue.NewQueueItem(fmt.Sprintf("csv-%v", time.Now().UnixNano()), map[string]interface{}{
					"cids": cids,
				}))

				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to queue"})
					return
				}
				cids = []string{}
			}
		}

		if len(cids) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing cids"})
			return
		}

		err = dynamodbQueue.AddItem(queue.NewQueueItem(fmt.Sprintf("csv-%v", time.Now().UnixNano()), map[string]interface{}{
			"cids": cids,
		}))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to queue"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	r.Run() // listen and serve on 0.0.0.0:8080

}
