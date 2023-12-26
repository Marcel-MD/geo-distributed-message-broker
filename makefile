up:
	docker-compose up --build

gen_proto:
	protoc --proto_path=api/proto api/proto/broker.proto --go_out=api --go-grpc_out=api

image:
	docker image tag geo-distributed-message-broker:latest marcelvlasenco/geo-distributed-message-broker:latest

push:
	docker image push marcelvlasenco/geo-distributed-message-broker:latest