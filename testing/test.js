import { check, sleep } from 'k6'
import grpc from 'k6/net/grpc'
import encoding from 'k6/encoding'
import { randomString, randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'

export const options = {
    scenarios: {
        publisher: {
            executor: 'ramping-arrival-rate',
            exec: 'publisher',
            preAllocatedVUs: 10,
            timeUnit: '1s',
            stages: [
                { target: 50, duration: '30s' },
                { target: 100, duration: '30s' },
                { target: 100, duration: '60s' },
                { target: 50, duration: '30s' },
                { target: 0, duration: '30s' },
            ]
        },
        subscriber: {
            executor: 'per-vu-iterations',
            exec: 'subscriber',
            vus: 10,
            iterations: 1
        },
    },
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)', 'count'],
    thresholds: {
        // Intentionally empty. We'll programatically define our bogus
        // thresholds (to generate the sub-metrics) below.
    }
};

for (let key in options.scenarios) {
    // Each scenario automaticall tags the metrics it generates with its own name
    let thresholdName = `grpc_req_duration{scenario:${key}}`;
    // Check to prevent us from overwriting a threshold that already exists
    if (!options.thresholds[thresholdName]) {
        options.thresholds[thresholdName] = [];
    }
    // 'max>=0' is a bogus condition that will always be fulfilled
    options.thresholds[thresholdName].push('max>=0');
}

const topics = ['Weather', 'Sports', 'Politics', 'Technology', 'Science'];
const nodes = ['localhost:8070', 'localhost:8080', 'localhost:8090']
const client = new grpc.Client();
client.load(['../proto'], 'broker.proto');

const params = {
    metadata: {
        'authorization': "basic " + encoding.b64encode('admin:password'),
    }
}

export function publisher() {
    client.connect(randomItem(nodes), {
        plaintext: true,
    });

    let request = {
        body: encoding.b64encode(randomString(20)),
        topic: randomItem(topics),
    };

    let response = client.invoke('broker.Broker/Publish', request, params);

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

    let stream = new grpc.Stream(client, 'broker.Broker/Subscribe', params);
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

    sleep(200);
    client.close();
}