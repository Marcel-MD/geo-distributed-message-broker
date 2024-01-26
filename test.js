import { check, sleep } from 'k6'
import grpc from 'k6/experimental/grpc'
import encoding from 'k6/encoding'
import { randomString, randomItem, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'

export const options = {
    scenarios: {
      publisher: {
        executor: 'per-vu-iterations',
        exec: 'publisher',
        vus: 1,
        iterations: 5,
        maxDuration: '30s',
      },
      subscriber: {
        executor: 'per-vu-iterations',
        exec: 'subscriber',
        vus: 5,
        iterations: 1,
        maxDuration: '30s',
      },
    },
  };

const topics = ['Alerts', 'News'];
const nodes = ['localhost:8070', 'localhost:8080', 'localhost:8090']
const client = new grpc.Client();
client.load(['proto'], 'broker.proto');

export function publisher() {
    sleep(randomIntBetween(1, 2));
    
    client.connect(randomItem(nodes), {
        plaintext: true,
    });

    let request = {
        body: encoding.b64encode(randomString(20)),
        topic: randomItem(topics),
    };

    let response = client.invoke('broker.Broker/Publish', request);

    check(response, {
        'publish is successful': (r) => r && r.status === grpc.StatusOK,
    });
    
    client.close();
}

export function subscriber() {
    client.connect(randomItem(nodes), {
        plaintext: true,
    });

    let request = {
        topics: {}
    };

    let timestamp = Date.now() * 1000;
    let topic = randomItem(topics);
    request.topics[topic] = timestamp;

    let stream = new grpc.Stream(client, 'broker.Broker/Subscribe');
    stream.write(request);
    stream.end();

    // sets up a handler for the data (server sends data) event 
    stream.on('data', (message) => {
        check(message, {
            'message is not empty': (m) => m && m.body !== '',
            'message has the right topic': (m) => m && m.topic === topic,
            'message has correct ordering': (m) => {
                if (!m) {
                    return false;
                }

                if (m.timestamp < timestamp) {
                    console.log(`Message ${m.id} has timestamp ${m.timestamp} which is less than ${timestamp}`)
                    return false;
                }

                timestamp = m.timestamp;
                return true;
            },
        })
    });

    // sets up a handler for the error event (an error occurs)
    stream.on('error', function (e) {
        check(e, {
            'error is null': (e) => !e || e.message === "canceled by client (k6)",
        });
    });
    
    sleep(15);
    client.close();
}