build:
	docker compose up --build

run:
	docker compose up

image:
	docker image tag geo-distributed-message-broker:latest marcelvlasenco/geo-distributed-message-broker:latest

push:
	docker image push marcelvlasenco/geo-distributed-message-broker:latest

gen_proto:
	protoc --proto_path=proto proto/broker.proto --go_out=. --go-grpc_out=.
	protoc --proto_path=proto proto/node.proto --go_out=. --go-grpc_out=.

k6:
	k6 run ./test.js

gen_cert:
	openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout cert/ca-key.pem -out cert/ca-cert.pem -subj "/C=MD/ST=Moldova/L=Chisinau/O=UTM/OU=FAF/CN=Marcel/emailAddress=marcel.vlasenco@isa.utm.md"
	openssl req -newkey rsa:4096 -nodes -keyout cert/server-key.pem -out cert/server-req.pem -subj "/C=MD/ST=Moldova/L=Chisinau/O=UTM/OU=FAF/CN=Marcel/emailAddress=marcel.vlasenco@isa.utm.md"
	openssl req -newkey rsa:4096 -nodes -keyout cert/client-key.pem -out cert/client-req.pem -subj "/C=MD/ST=Moldova/L=Chisinau/O=UTM/OU=FAF/CN=Marcel/emailAddress=marcel.vlasenco@isa.utm.md"

sign_cert:
	openssl x509 -req -in cert/server-req.pem -CA cert/ca-cert.pem -CAkey cert/ca-key.pem -CAcreateserial -out cert/server-cert.pem -extfile cert/server-ext.conf
	openssl x509 -req -in cert/client-req.pem -CA cert/ca-cert.pem -CAkey cert/ca-key.pem -CAcreateserial -out cert/client-cert.pem -extfile cert/client-ext.conf