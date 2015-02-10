all: sqs-irc-gateway

sqs-irc-gateway:
	go build

clean:
	rm -f sqs-irc-gateway
