## Functional Requirements

## Non-Functional Requirements

| Identifier | Requirement | Target Value |
| --- | --- | --- |
| NFR 01 | Response Time | The average response time for message publication and delivery remains below a defined threshold of 15 milliseconds under typical load conditions. |
| NFR 02 | Throughput | The system should support a minimum message throughput of 1,000 msgs/sec, with provisions for scaling to 10,000 msgs/sec as needed. |
| NFR 03 | Latency | Latency with no more than 10 milliseconds delay between message publishing and delivery, even under peak loads. |
| NFR 04 | Scalability | Linear horizontal scalability: Proportional increase with new nodes across different geographical regions. Vertical scalability: Support hardware upgrades for nodes in various regions. |
| NFR 05 | Resource Utilization | Keep CPU usage below 70% and memory usage below 80% under peak loads. Maintain network bandwidth usage below 70% of the available capacity. |
| NFR 06 | Fault Tolerance | Ensure that the message broker can continue to operate with minimal performance degradation even in the presence of node failures or network interruptions. |
| NFR 07 | High Availability | Aim for at least 99.99% uptime over a year (approximately 52 minutes of downtime per year). |
| NFR 08 | Consistency and Data Integrity | Guarantee strong consistency and data integrity across all nodes, ensuring that messages are delivered in the order they were published. Replication factor of at least 3 for data consistency. |
| NFR 09 | Security | Implement performance measures to protect against unauthorized access, DDoS attacks, and data breaches, while maintaining minimal impact on system performance. |
| NFR 10 | Load Balancing | Ensure clients can publish or consume messages from nodes near their geographical area. |
| NFR 11 | Failure Recovery | Define a recovery time objective (RTO) of under 5 minutes for system recovery after failures. |
| NFR 12 | Message Retention and Cleanup | Configurable message retention with automatic cleanup. |
| NFR 13 | Message Delivery Guarantee | Guarantee at-least-once delivery of messages. |
| NFR 14 | Message Ordering Guarantee | Guarantee in-order delivery of messages. |
| NFR 15 | Message Durability | Guarantee message durability in the event of node failures. |
| NFR 16 | Interface | Provide gRPC interface for clients to publish and consume messages. |
| NFR 18 | Logging | Support logging for system administrators to monitor the health of the system. |
| NFR 19 | Deployment | Provide a deployment mechanism to deploy the system on a cloud platform. |

## Architecture

![Developer use case diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/use_dev.png)
![Consumer use case diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/use_cons.png)
![Producer use case diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/use_prod.png)

![Nodes sequence diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/seq_nodes.png)
![Conflict sequence diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/seq_conflict.png)

![Consumer activity diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/act_consume.png)
![Deployment diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/depl.png)
![Component diagram](https://github.com/Marcel-MD/thesis-uml/blob/main/comp.png)