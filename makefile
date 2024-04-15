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

k6:
	k6 run ./testing/test.js

k6_prometheus:
	K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	K6_PROMETHEUS_RW_TREND_AS_NATIVE_HISTOGRAM=true \
	k6 run -o experimental-prometheus-rw ./testing/test.js