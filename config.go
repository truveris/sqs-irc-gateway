// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
	"errors"

	"github.com/jessevdk/go-flags"
)

type Cmd struct {
	ConfigFile string `short:"c" description:"Configuration file" default:"/etc/sqs-irc-gateway.conf"`
}

type Cfg struct {
	// Credentials used to write to the minion queues and read from the
	// soul queue.
	AWSAccessKeyId     string
	AWSSecretAccessKey string

	// AWS Region to use for SQS access (e.g. us-east-1).
	AWSRegionCode string

	// Server name as expected by Go's Dial command, it should contain the
	// port number (e.g. localhost:6667).
	IRCServer string

	// Name of the bot upon startup. This can naturally be changed with
	// commands coming from the outgoing queue but this particular
	// sub-system should not care, this is only for initialization.
	IRCNickname string

	// These are incoming and outgoing from the perspective of this process
	// toward the IRC server (incoming from IRC, outgoing to IRC). Which
	// means the AWS credentials above should give the right access:
	// GetMessages on the outgoing queue, and SendMessage on the incoming
	// queue.
	IncomingQueueName string
	OutgoingQueueName string
}

var (
	cmd = Cmd{}
	cfg = Cfg{}
)

// Parse the command line arguments and return the soul program's path/name
// (only argument).
func parseCommandLine() {
	flagParser := flags.NewParser(&cmd, flags.PassDoubleDash)
	_, err := flagParser.Parse()
	if err != nil {
		println("command line error: " + err.Error())
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
}

// Look in the current directory for an config.json file.
func parseConfigFile() error {
	file, err := os.Open(cmd.ConfigFile)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	if cfg.IRCNickname == "" {
		return errors.New("'IRCNickname' is not defined")
	}

	if cfg.IncomingQueueName == "" {
		return errors.New("'IRCIncomingQueueName' is not defined")
	}

	if cfg.OutgoingQueueName == "" {
		return errors.New("'IRCOutgoingQueueName' is not defined")
	}

	if cfg.AWSRegionCode == "" {
		return errors.New("'AWSRegionCode' is not defined")
	}

	if cfg.AWSAccessKeyId == "" {
		return errors.New("'AWSAccessKeyId' is not defined")
	}

	if cfg.AWSSecretAccessKey == "" {
		return errors.New("'AWSSecretAccessKey' is not defined")
	}

	return nil
}
