// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"time"

	"github.com/truveris/sqs"
)

// Feeds the given queue with messages fetched from the given SQS queue.
func sqsReader(client *sqs.Client, queue chan string, errors chan error) {
	url, err := client.CreateQueue(cfg.IncomingQueueName)
	if err != nil {
		errors <- err
		return
	}

	for {
		req := sqs.NewReceiveMessageRequest(url)
		req.Set("AttributeName", "SenderId")
		req.Set("WaitTimeSeconds", "20")

		msg, err := client.GetSingleMessageFromRequest(req)
		if err != nil {
			log.Printf("sqsReader GetMessage error: %s", err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		if msg == nil {
			continue
		}

		queue <- msg.Body
	}
}

func sqsWriter(client *sqs.Client, queue chan string, errors chan error) {
	url, err := client.CreateQueue(cfg.OutgoingQueueName)
	if err != nil {
		errors <- err
		return
	}

	for {
		err = client.SendMessage(url, <-queue)
		if err != nil {
			log.Printf("sqsWriter SendMessage error: %s", err.Error())
			continue
		}
	}
}
