import { check, sleep } from 'k6'
import grpc from 'k6/net/grpc'
import encoding from 'k6/encoding'
import { randomString, randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'

export const options = {
    stages: [
        { duration: '30s', target: 10 }
    ],
};

const topics = ['Alerts', 'News'];
const nodes = ['localhost:8070', 'localhost:8080', 'localhost:8090']

const client = new grpc.Client();
client.load(['proto'], 'broker.proto');

export default () => {
    client.connect(randomItem(nodes), {
        plaintext: true,
    });

    const data = {
        body: encoding.b64encode(randomString(20)),
        topic: randomItem(topics),
    };

    const response = client.invoke('broker.Broker/Publish', data);

    check(response, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });
    
    client.close();
    sleep(1);
}
