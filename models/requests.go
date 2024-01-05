package models

import (
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/pb"
)

type ProposeRequest struct {
	Message data.Message
}

func (r ProposeRequest) ToPb() *pb.ProposeRequest {
	return &pb.ProposeRequest{
		Message: messageToPb(r.Message),
	}
}

func ToProposeRequest(rsp *pb.ProposeRequest) ProposeRequest {
	return ProposeRequest{
		Message: messageFromPb(rsp.Message),
	}
}

type ProposeResponse struct {
	Ack          bool
	Message      data.Message
	Predecessors map[string]data.Message
}

func (r ProposeResponse) ToPb() *pb.ProposeResponse {
	return &pb.ProposeResponse{
		Ack:          r.Ack,
		Message:      messageToPb(r.Message),
		Predecessors: messagesToPb(r.Predecessors),
	}
}

func ToProposeResponse(rsp *pb.ProposeResponse) ProposeResponse {
	return ProposeResponse{
		Ack:          rsp.Ack,
		Message:      messageFromPb(rsp.Message),
		Predecessors: messagesFromPb(rsp.Predecessors),
	}
}

type StableRequest struct {
	Message      data.Message
	Predecessors map[string]data.Message
}

func (r StableRequest) ToPb() *pb.StableRequest {
	return &pb.StableRequest{
		Message:      messageToPb(r.Message),
		Predecessors: messagesToPb(r.Predecessors),
	}
}

func ToStableRequest(rsp *pb.StableRequest) StableRequest {
	return StableRequest{
		Message:      messageFromPb(rsp.Message),
		Predecessors: messagesFromPb(rsp.Predecessors),
	}
}

func messageToPb(msg data.Message) *pb.Message {
	return &pb.Message{
		Id:        msg.ID,
		Timestamp: msg.Timestamp,
		Topic:     msg.Topic,
		Body:      msg.Body,
	}
}

func messagesToPb(msgs map[string]data.Message) map[string]*pb.Message {
	messages := make(map[string]*pb.Message)

	for _, msg := range msgs {
		messages[msg.ID] = messageToPb(msg)
	}

	return messages
}

func messageFromPb(msg *pb.Message) data.Message {
	return data.Message{
		ID:        msg.Id,
		Timestamp: msg.Timestamp,
		Topic:     msg.Topic,
		Body:      msg.Body,
	}
}

func messagesFromPb(msgs map[string]*pb.Message) map[string]data.Message {
	messages := make(map[string]data.Message)

	for _, msg := range msgs {
		messages[msg.Id] = messageFromPb(msg)
	}

	return messages
}
