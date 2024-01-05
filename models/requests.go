package models

import (
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/proto"
)

type ProposeRequest struct {
	Message data.Message
}

func (r ProposeRequest) ToProto() *proto.ProposeRequest {
	return &proto.ProposeRequest{
		Message: messageToProto(r.Message),
	}
}

func ToProposeRequest(rsp *proto.ProposeRequest) ProposeRequest {
	return ProposeRequest{
		Message: messageFromProto(rsp.Message),
	}
}

type ProposeResponse struct {
	Ack          bool
	Message      data.Message
	Predecessors map[string]data.Message
}

func (r ProposeResponse) ToProto() *proto.ProposeResponse {
	return &proto.ProposeResponse{
		Ack:          r.Ack,
		Message:      messageToProto(r.Message),
		Predecessors: messagesToProto(r.Predecessors),
	}
}

func ToProposeResponse(rsp *proto.ProposeResponse) ProposeResponse {
	return ProposeResponse{
		Ack:          rsp.Ack,
		Message:      messageFromProto(rsp.Message),
		Predecessors: messagesFromProto(rsp.Predecessors),
	}
}

type StableRequest struct {
	Message      data.Message
	Predecessors map[string]data.Message
}

func (r StableRequest) ToProto() *proto.StableRequest {
	return &proto.StableRequest{
		Message:      messageToProto(r.Message),
		Predecessors: messagesToProto(r.Predecessors),
	}
}

func ToStableRequest(rsp *proto.StableRequest) StableRequest {
	return StableRequest{
		Message:      messageFromProto(rsp.Message),
		Predecessors: messagesFromProto(rsp.Predecessors),
	}
}

func messageToProto(msg data.Message) *proto.Message {
	return &proto.Message{
		Id:        msg.ID,
		Timestamp: msg.Timestamp,
		Topic:     msg.Topic,
		Body:      msg.Body,
	}
}

func messagesToProto(msgs map[string]data.Message) map[string]*proto.Message {
	messages := make(map[string]*proto.Message)

	for _, msg := range msgs {
		messages[msg.ID] = messageToProto(msg)
	}

	return messages
}

func messageFromProto(msg *proto.Message) data.Message {
	return data.Message{
		ID:        msg.Id,
		Timestamp: msg.Timestamp,
		Topic:     msg.Topic,
		Body:      msg.Body,
	}
}

func messagesFromProto(msgs map[string]*proto.Message) map[string]data.Message {
	messages := make(map[string]data.Message)

	for _, msg := range msgs {
		messages[msg.Id] = messageFromProto(msg)
	}

	return messages
}
