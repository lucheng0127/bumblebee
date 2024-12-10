package service

import (
	"net"
	"sync"

	"github.com/lucheng0127/bumblebee/pkg/dispatch/core"
	"github.com/nats-io/nats.go"
)

func NewClientLinstener(topic string, conn *nats.Conn) net.Listener {
	return &core.NatsLinstener{
		LocalTopic: topic,
		ConnMap:    make(map[string]*core.NatsConn),
		Lock:       sync.Mutex{},
		Closed:     false,
		Nats:       conn,
	}
}
