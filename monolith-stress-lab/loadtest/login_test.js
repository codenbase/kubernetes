import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';

let errorCount = 0;

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS) : 200,
  duration: __ENV.DURATION || '30s',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export function setup() {
  const cost = __ENV.COST ? parseInt(__ENV.COST) : null;
  const baseUrl = __ENV.BASE_URL || 'http://app:8080';
  
  if (cost) {
    const payload = JSON.stringify({ cost: cost });
    const params = { headers: { 'Content-Type': 'application/json' } };
    const res = http.put(`${baseUrl}/config/bcrypt-cost`, payload, params);
    
    if (res.status === 200) {
      console.log(`[*] Setup: Auto-tuning Bcrypt Cost to ${cost} SUCCESS`);
    } else {
      console.error(`[!] Setup: Failed to tune Bcrypt Cost to ${cost}: ${res.status} ${res.body}`);
    }
  }
}

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
  
  const baseUrl = __ENV.BASE_URL || 'http://app:8080';
  const res = http.post(`${baseUrl}/login`, payload, params);

  if (res.status !== 200) {
    if (errorCount < 3) {
      console.error(`[VU ${exec.vu.idInTest}] Login Failed. Status: ${res.status}, Error: ${res.error}, Body: ${res.body}`);
      errorCount++;
    }
  }
  
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
