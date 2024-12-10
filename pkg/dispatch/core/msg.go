package core

import "github.com/vmihailenco/msgpack"

type DispatchMsg struct {
	SenderTopic string
	TargetTopic string
	OperationID string
	FinMsg      bool
	Payload     []byte
}

func (msg *DispatchMsg) Encode() ([]byte, error) {
	return msgpack.Marshal(msg)
}

func Decode(data []byte) (*DispatchMsg, error) {
	msg := new(DispatchMsg)

	err := msgpack.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
