# Geo Distributed Message Broker

To run the message broker in it's default configuration:
```bash
docker-compose up --build
```

Message broker nodes will be available on `localhost:8070`, `localhost:8080`, and `localhost:8090`.  
Interact with them using gRPC client of your choice:
```proto
syntax = "proto3";

package broker;

service Broker {
    rpc Publish (PublishRequest) returns (PublishResponse);
    rpc Subscribe(SubscribeRequest) returns (stream MessageResponse);
}

message PublishRequest {
    string topic = 1;
    bytes body = 2;
}

message PublishResponse {
    string id = 1;
}

message SubscribeRequest {
    repeated string topics = 1;
    string consumer_name = 2;
}

message MessageResponse {
    string id = 1;
    int64 timestamp = 2;
    string topic = 3;
    bytes body = 4;
}
```

### Thesis Topic
Cloud-native fault-tolerant message broker, part of toolchain for geo-distributed systems

### Subject
Design and implement a distributed, fault-tolerant message broker system that leverages consensus algorithms for ensuring reliability and consistency across a wide area network. This project aims to address the needs of organizations requiring a highly available and dependable messaging infrastructure within the scope of geographically distributed systems.

### Objectives
- Ensure the message broker reliably handles message publishing and delivery under various conditions.
- Incorporate a consensus algorithm for managing distributed state and ensuring data consistency among nodes.
- Develop native support for gRPC communication to enable efficient and secure inter-service interactions.
- Design the system to be horizontally scalable to accommodate increased message volumes and evolving infrastructure needs.
- Create comprehensive documentation outlining system architecture, configuration, and deployment guidelines for users.
- Implement security measures to protect the system from unauthorized access, ensuring data confidentiality and integrity.

## Requirements

### Consumer Requirements
1. The system must provide authentication mechanisms for consumers, ensuring secure access to
topics.
2. Consumers should have the ability to subscribe to specific topics of interest, allowing them to
receive relevant messages.
3. Consumers must be able to unsubscribe from topics when they no longer wish to receive
messages from those topics.
4. Consumers should be able to consume messages from the subscribed topics, enabling them to
process the messages as needed.
5. Consumers should be able to receive messages from the nearest data center in the geo-distributed
network to minimize latency and enhance the real-time nature of their applications.

![Consumer use case diagram](./diagrams/use_cons.png)

### Producer Requirements
1. The system must provide authentication mechanisms for producers, ensuring secure publishing
of messages to topics.
2. Producers must be able to create new topics in the message broker for effective message categorization and publishing.
3. Producers should be able to publish messages to specific topics, with the message broker
ensuring reliable message delivery to consumers.
4. Producers should be able to publish messages to the nearest data center in the geo-distributed
network to minimize latency.

![Producer use case diagram](./diagrams/use_prod.png)

### Developer Requirements
1. Developers should have access to comprehensive documentation that explains the message
broker’s features, capabilities, and usage.
2. Developers must be able to configure the message broker with various settings, including security, replication, and geo-distribution options.
3. Developers should be able to deploy the message broker using a Docker image, simplifying
scalability and management across different environments.
4. Developers should be able to manage the scalability of the message broker network by dynamically adding or removing nodes based on traffic demands.

![Developer use case diagram](./diagrams/use_dev.png)

## Architecture

### Consensus No Conflicts
![Nodes sequence diagram](./diagrams/seq_nodes.png)

### Consensus With Conflicts
![Conflict sequence diagram](./diagrams/seq_conflict.png)

### Deployment Across Multiple Regions
![Deployment diagram](./diagrams/depl.png)

### Component Diagram
![Component diagram](./diagrams/comp.png)