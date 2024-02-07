package persistencepg

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// struct
type envelope struct {
	ID         string
	Message    json.RawMessage
	EventIndex int
	ActorName  string
	EventName  string
}

type messageEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func newEnvelope(message proto.Message) ([]byte, error) {
	typeName := proto.MessageName(message)
	bytes, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&messageEnvelope{Payload: bytes, Type: string(typeName)})
}

func (envelope *envelope) message() (proto.Message, error) {
	me := &messageEnvelope{}
	err := json.Unmarshal(envelope.Message, me)
	if err != nil {
		return nil, err
	}
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(me.Type))
	if err != nil {
		return nil, err
	}
	pm := mt.New().Interface()
	err = json.Unmarshal(me.Payload, pm)
	if err != nil {
		return nil, err
	}
	return pm, nil
}
