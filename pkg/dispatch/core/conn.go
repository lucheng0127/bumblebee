package core

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type NatsConn struct {
	Nats        *nats.Conn
	PayloadChan chan []byte
	CloseFunc   func()
	LocalTopic  string
	RemoteTopic string
	OperationID string
	Closed      bool
}

func (conn *NatsConn) Read(b []byte) (int, error) {
	if conn.Closed {
		return 0, errors.New("can't read from a closed nats conn")
	}

	payload := <-conn.PayloadChan
	n := copy(b, payload)

	log.Debugf("read [%d] bytes from nats with rtopic [%s] ltopic [%s] operation id [%s]", n, conn.RemoteTopic, conn.LocalTopic, conn.OperationID)
	return n, nil
}

func (conn *NatsConn) Write(b []byte) (int, error) {
	if conn.Closed {
		return 0, errors.New("can't write to a closed nats conn")
	}

	n := len(b)
	dMsg := DispatchMsg{
		SenderTopic: conn.LocalTopic,
		TargetTopic: conn.RemoteTopic,
		OperationID: conn.OperationID,
		FinMsg:      false,
		Payload:     b,
	}

	dMsgBytes, err := dMsg.Encode()
	if err != nil {
		return 0, fmt.Errorf("failed to encode dispatch msg: %s", err.Error())
	}

	err = conn.Nats.Publish(conn.RemoteTopic, dMsgBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to send nats msg with topic %s: %s", conn.RemoteTopic, err.Error())
	}

	log.Debugf("write [%d] bytes to nats with rtopic [%s] ltopic [%s] operation id [%s]", n, conn.RemoteTopic, conn.LocalTopic, conn.OperationID)
	return n, nil
}

func (conn *NatsConn) Close() error {
	dMsg := DispatchMsg{
		SenderTopic: conn.LocalTopic,
		TargetTopic: conn.RemoteTopic,
		OperationID: conn.OperationID,
		FinMsg:      true,
		Payload:     []byte("final"),
	}

	dMsgBytes, err := dMsg.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode final dispatch msg: %s", err.Error())
	}

	err = conn.Nats.Publish(conn.RemoteTopic, dMsgBytes)
	if err != nil {
		return fmt.Errorf("failed to send final nats msg with topic %s: %s", conn.RemoteTopic, err.Error())
	}

	log.Debugf("send fin msg to nats with rtopic [%s] ltopic [%s] operation id [%s]", conn.RemoteTopic, conn.LocalTopic, conn.OperationID)
	conn.Closed = true
	conn.CloseFunc()

	return nil
}

type NatAddr struct {
	topic string
}

func (addr *NatAddr) Network() string {
	return "nats"
}

func (addr *NatAddr) String() string {
	return fmt.Sprintf("topic-%s", addr.topic)
}

func (conn *NatsConn) LocalAddr() net.Addr {
	return &NatAddr{conn.LocalTopic}
}

func (conn *NatsConn) RemoteAddr() net.Addr {
	return &NatAddr{conn.RemoteTopic}
}

func (conn *NatsConn) SetDeadline(time.Time) error {
	return nil
}

func (conn *NatsConn) SetReadDeadline(time.Time) error {
	return nil
}

func (conn *NatsConn) SetWriteDeadline(time.Time) error {
	return nil
}
