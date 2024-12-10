package core

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type NatsLinstener struct {
	LocalTopic string
	ConnMap    map[string]*NatsConn
	Lock       sync.Mutex
	Closed     bool
	Nats       *nats.Conn
}

func (ln *NatsLinstener) Close() error {
	ln.Closed = true

	ln.Lock.Lock()
	for k, _ := range ln.ConnMap {
		delete(ln.ConnMap, k)
	}
	ln.Lock.Unlock()

	return nil
}

func (ln *NatsLinstener) Addr() net.Addr {
	return &NatAddr{ln.LocalTopic}
}

func (ln *NatsLinstener) getOrCreateConn(msg *DispatchMsg) *NatsConn {
	key := msg.OperationID
	conn, ok := ln.ConnMap[key]
	if ok {
		return conn
	}

	conn = &NatsConn{
		Nats:        ln.Nats,
		PayloadChan: make(chan []byte, 1024),
		LocalTopic:  msg.TargetTopic,
		RemoteTopic: msg.SenderTopic,
		OperationID: msg.OperationID,
		Closed:      false,
		CloseFunc: func() {
			ln.Lock.Lock()
			delete(ln.ConnMap, key)
			ln.Lock.Unlock()
		},
	}

	ln.Lock.Lock()
	ln.ConnMap[key] = conn
	ln.Lock.Unlock()
	return conn
}

func (ln *NatsLinstener) Accept() (net.Conn, error) {
	if ln.Closed {
		return nil, errors.New("can't accept from a closed nats listener")
	}

	msgChan := make(chan *DispatchMsg, 1024)
	errMsg := ""

	_, err := ln.Nats.Subscribe(ln.LocalTopic, func(msg *nats.Msg) {
		dMsg, err := Decode(msg.Data)
		if err != nil {
			errMsg = err.Error()
			return
		}

		key := dMsg.OperationID

		if dMsg.FinMsg {
			_, ok := ln.ConnMap[key]
			if !ok {
				return
			}

			// If final msg, try to delete from conn map and do nothing
			ln.Lock.Lock()
			delete(ln.ConnMap, key)
			ln.Lock.Unlock()

			log.Debugf("fin msg received from nats with rtopic [%s] ltopic [%s] operation id [%s]", dMsg.SenderTopic, dMsg.TargetTopic, dMsg.OperationID)
			return
		}

		msgChan <- dMsg
	})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe topic %s from nats: %s", ln.LocalTopic, err.Error())
	}

	if errMsg != "" {
		return nil, fmt.Errorf("failed to decode dispatch msg: %s", errMsg)
	}

	// Block until message received
	msg := <-msgChan
	conn := ln.getOrCreateConn(msg)

	conn.PayloadChan <- msg.Payload
	return conn, nil
}
