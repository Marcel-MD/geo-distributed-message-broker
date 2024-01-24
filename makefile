up:
	docker-compose up --build

gen_proto:
	protoc --proto_path=proto proto/broker.proto --go_out=. --go-grpc_out=.
	protoc --proto_path=proto proto/node.proto --go_out=. --go-grpc_out=.

image:
	docker image tag geo-distributed-message-broker:latest marcelvlasenco/geo-distributed-message-broker:latest

push:
	docker image push marcelvlasenco/geo-distributed-message-broker:latest

k6:
	k6 run ./test.js