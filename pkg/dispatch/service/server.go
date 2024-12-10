package service

import (
	"context"
	"net"
	"sync"

	"github.com/lucheng0127/bumblebee/pkg/dispatch/core"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type DispatchServer struct {
	LocalTopic string
	ConnMap    map[string]*core.NatsConn
	Lock       sync.Mutex
	Nats       *nats.Conn
}

func NewDispatchServer(topic string, conn *nats.Conn) *DispatchServer {
	return &DispatchServer{
		LocalTopic: topic,
		ConnMap:    make(map[string]*core.NatsConn),
		Lock:       sync.Mutex{},
		Nats:       conn,
	}
}

func (svc *DispatchServer) NewNatsConn(topic, operationId string) net.Conn {
	key := operationId
	conn := &core.NatsConn{
		Nats:        svc.Nats,
		PayloadChan: make(chan []byte, 1024),
		LocalTopic:  svc.LocalTopic,
		RemoteTopic: topic,
		OperationID: operationId,
		Closed:      false,
		CloseFunc: func() {
			svc.Lock.Lock()
			delete(svc.ConnMap, key)
			svc.Lock.Unlock()
		},
	}

	svc.Lock.Lock()
	svc.ConnMap[key] = conn
	svc.Lock.Unlock()
	return conn
}

func (svc *DispatchServer) Run(ctx context.Context, exitFunc func()) {
	_, err := svc.Nats.Subscribe(svc.LocalTopic, func(msg *nats.Msg) {
		dMsg, err := core.Decode(msg.Data)
		if err != nil {
			log.Errorf("failed to decode dispatch msg: %s", err.Error())
			return
		}
		key := dMsg.OperationID

		// Check fin msg
		if dMsg.FinMsg {
			// If final msg, try to delete from conn map and do nothing
			svc.Lock.Lock()
			delete(svc.ConnMap, key)
			svc.Lock.Unlock()

			log.Debugf("fin msg received from nats with rtopic [%s] ltopic [%s] operation id [%s]", dMsg.SenderTopic, dMsg.TargetTopic, dMsg.OperationID)
			return
		}

		conn, ok := svc.ConnMap[key]
		if !ok {
			log.Warnf("failed to get nats connection for topic %s", key)
			return
		}

		if conn.Closed {
			svc.Lock.Lock()
			delete(svc.ConnMap, key)
			svc.Lock.Unlock()

			log.Warnf("nats connection closed with topic %s", key)
			return
		}

		conn.PayloadChan <- dMsg.Payload
	})

	if err != nil {
		log.Errorf("failed to subscribe topic from nats with topic %s: %s", svc.LocalTopic, err.Error())
		exitFunc()
	}

	select {
	case <-ctx.Done():
		log.Infof("exit signal received, stop subsribe %s from nats", svc.LocalTopic)
		exitFunc()
	}
}
