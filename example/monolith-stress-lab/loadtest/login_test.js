import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS) : 200,
  duration: __ENV.DURATION || '30s',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export default function () {
  const userIndex = Math.floor(Math.random() * 10000) + 1;
  const username = `user${userIndex}`;
  const password = '123456';
  
  const payload = JSON.stringify({
    username: username,
    password: password,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const res = http.post('http://app:8080/login', payload, params);
  
  check(res, {
    'is status 200': (r) => r.status === 200,
    'has token': (r) => {
        try {
            const body = r.json();
            return body.token !== undefined && body.token !== '';
        } catch (e) {
            return false;
        }
    }
  });
}
