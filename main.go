// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// Here is a shitty diagram to explain how this work.
//
//           +------------------+       +-----------------+
//           | IRC Socket Write |       | IRC Socket Read |
//           +----------------_-+       +-_---------------+
//                            |           |
//              /|\     +-----^---+   +---^----+       |
//               |      | irc out |   | irc in |       |
//               |      +-----_---+   +---_----+       |
//               |            |           |            |
//               |            |           |            |
//        to irc |         +--^-----------^--+         | from irc
//               |         | sqs-irc-gateway |         |
//               |         +--_-----------_--+         |
//               |            |           |            |
//               |            |           |            |
//               |       +----^---+   +---^-----+      |
//               |       | sqs in |   | sqs out |     \|/
//                       +----_---+   +---_-----+
//                            |           |
//         +------------------^-+       +-^------------------+
//         | IRC Outgoing Queue |       | IRC Incoming Queue |
//         +--------------------+       +--------------------+
//
// The program using these queues just need to read the incoming queue and
// write to the outgoing queue.
//

package main

import (
	"errors"
	"log"
	"os"
	"time"

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
	client, err := sqs.NewClient(cfg.AWSAccessKeyId, cfg.AWSSecretAccessKey,
		cfg.RegionCode)
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
	// queues (from the perspective of this program). The incoming channel
	// for SQS is hooked up to the outgoing IRC queue (this process receives
	// all the IRC messages going back to the server).
	sqsin, sqsinerrch, err := sqschan.Incoming(client, cfg.OutgoingQueueName)
	if err != nil {
		return errors.New("unable to create sqschan.Incoming: " +
			err.Error())
	}
	sqsout, sqsouterrch, err := sqschan.Outgoing(client, cfg.IncomingQueueName)
	if err != nil {
		return errors.New("unable to create sqschan.Outcoming: " +
			err.Error())
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
			log.Printf("SQS Error on incoming channel: %s",
				err.Error())
			time.sleep(10 * time.Second)
		case err = <-sqsouterrch:
			log.Printf("SQS Error on outgoing channel: %s",
				err.Error())
			time.sleep(10 * time.Second)
		}
	}

	return nil
}

func main() {
	parseCommandLine()
	parseConfigFile()

	log.Printf("starting %s", cfg.Nickname)
	err := start()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
