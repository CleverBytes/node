package nats

import (
	"github.com/mysterium/node/communication"
	"github.com/nats-io/go-nats"
	"time"

	"fmt"
	log "github.com/cihub/seelog"
)

const SENDER_LOG_PREFIX = "[NATS.Sender] "

type senderNats struct {
	connection     *nats.Conn
	codec          communication.Codec
	timeoutRequest time.Duration
	messageTopic   string
}

func (sender *senderNats) Send(packer communication.MessagePacker) error {

	messageType := string(packer.GetMessageType())

	messageData, err := sender.codec.Pack(packer.CreateMessage())
	if err != nil {
		err = fmt.Errorf("Failed to encode message '%s'. %s", messageType, err)
		return err
	}

	log.Debug(SENDER_LOG_PREFIX, fmt.Sprintf("Message '%s' sending: %s", messageType, messageData))
	err = sender.connection.Publish(
		sender.messageTopic+messageType,
		messageData,
	)
	if err != nil {
		err = fmt.Errorf("Failed to send message '%s'. %s", messageType, err)
		return err
	}

	return nil
}

func (sender *senderNats) Request(packer communication.RequestPacker) (responsePtr interface{}, err error) {

	requestType := string(packer.GetRequestType())
	responsePtr = packer.CreateResponse()

	requestData, err := sender.codec.Pack(packer.CreateRequest())
	if err != nil {
		err = fmt.Errorf("Failed to pack request '%s'. %s", requestType, err)
		return
	}

	log.Debug(SENDER_LOG_PREFIX, fmt.Sprintf("Request '%s' sending: %s", requestType, requestData))
	msg, err := sender.connection.Request(
		sender.messageTopic+requestType,
		requestData,
		sender.timeoutRequest,
	)
	if err != nil {
		err = fmt.Errorf("Failed to send request '%s'. %s", requestType, err)
		return
	}

	log.Debug(SENDER_LOG_PREFIX, fmt.Sprintf("Received response for '%s': %s", requestType, msg.Data))
	err = sender.codec.Unpack(msg.Data, responsePtr)
	if err != nil {
		err = fmt.Errorf("Failed to unpack response '%s'. %s", requestType, err)
		log.Error(RECEIVER_LOG_PREFIX, err)
		return
	}

	return responsePtr, nil
}
