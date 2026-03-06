import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS) : 200,
  duration: __ENV.DURATION || '30s',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

// Setup runs once at the beginning of the test.
// We use it to login ONE time and retrieve a JWT token
// to simulate the scenario where a user repeatedly accesses protected APIs.
export function setup() {
  const payload = JSON.stringify({
    username: 'user1',
    password: '123456',
  });
  
  const params = {
    headers: { 'Content-Type': 'application/json' },
  };
  
  // Notice we use the internal docker network hostname 'app' here
  const res = http.post('http://app:8080/login', payload, params);
  
  let token = "";
  if (res.status === 200) {
      token = res.json('token');
  }
  return { token: token };
}

// The default function runs repeatedly by the VUs.
export default function (data) {
  const params = {
    headers: {
      'Authorization': `Bearer ${data.token}`
    },
  };
  
  // Fire requests at the fast JWT verification endpoint
  const res = http.get('http://app:8080/verify', params);
  
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}
