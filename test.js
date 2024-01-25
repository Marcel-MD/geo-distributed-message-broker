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
    client.connect(randomItem(nodes), {
        plaintext: true,
    });

    let request = {
        body: encoding.b64encode(randomString(20)),
        topic: randomItem(topics),
    };

    let response = client.invoke('broker.Broker/Publish', request);

    check(response, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });
    
    sleep(randomIntBetween(1, 2));
    client.close();
}

export function subscriber() {
    client.connect(randomItem(nodes), {
        plaintext: true,
    });

    let request = {
        topics: {}
    };

    let topic = randomItem(topics);
    request.topics[topic] = 0;

    let stream = new grpc.Stream(client, 'broker.Broker/Subscribe');
    stream.write(request);

    // sets up a handler for the data (server sends data) event 
    stream.on('data', (stats) => {
        console.log('Received', stats);
    });
    
    // sets up a handler for the end event (stream closes)
    stream.on('end', function () {
        // The server has finished sending
        client.close();
        console.log('All done');
    });

    // sets up a handler for the error event (an error occurs)
    stream.on('error', function (e) {
        // An error has occurred and the stream has been closed.
        console.log('Error: ' + JSON.stringify(e));
    });

    stream.end();
    
    sleep(30);
    client.close();
}