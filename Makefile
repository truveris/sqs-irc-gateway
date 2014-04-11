all: sqs-irc-gateway

ygor:
	go build

clean:
	rm -f sqs-irc-gateway
