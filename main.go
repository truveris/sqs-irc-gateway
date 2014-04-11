// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
)

func start() error {
	// Start IRC connection:
	conn, err := connect()
	if err != nil {
		return err
	}

	// Start SQS connection:
	client, err := sqs.NewClient(cfg.AWSSecretAccessKey,
		cfg.AWSAccessKeyId, cfg.RegionCode)
	if err != nil {
		return err
	}

	// These channels represent the lines coming and going to the IRC
	// server (from the perspective of this program).
	ircin := make(chan string)
	ircout := make(chan string)
	ircdisc := make(chan string)
	go connectionReader(conn, ircin, ircdisc)
	go connectionWriter(conn, ircout)

	// These channels represent the lines coming and going to the SQS
	// queues (from the perspective of this program).
	sqsin, sqsinerrch, err := sqschan.Incoming(client, cfg.IncomingQueueName)
	if err != nil {
		return err
	}
	sqsout, sqsouterrch, err := sqschan.Outgoing(client, cfg.OutgoingQueueName)
	if err != nil {
		return err
	}

	for {
		select {
		case data := <-ircin:
			// SQS <- IRC
			sqsout <- data
		case msg := <-sqsin:
			// IRC <- SQS 
			ircout <- msg.Body
			client.DeleteMessage(msg.QueueURL, msg.ReceiptHandle)
		case data := <-ircdisc:
			// Server has disconnected, we're done.
			log.Printf("Disconnected: %s", data)
			return nil
		case err = <-sqsinerrch:
			log.Printf("Fatal SQS Error on incoming channel: %s",
				err.Error())
			return nil
		case err = <-sqsouterrch:
			log.Printf("Fatal SQS Error on outgoing channel: %s",
				err.Error())
			return nil
		}
	}

	return nil
}

func main() {
	parseCommandLine()

	log.Printf("starting %s", cfg.Nickname)
	err := start()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
