import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';

let errorCount = 0;

export const options = {
  vus: __ENV.VUS ? parseInt(__ENV.VUS) : 200,
  duration: __ENV.DURATION || '30s',
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export default function () {
  // Random large article ID 101-200
  const id = Math.floor(Math.random() * 100) + 101;
  
  const baseUrl = __ENV.BASE_URL || 'http://app:8080';
  const res = http.get(`${baseUrl}/articles/${id}`);

  if (res.status !== 200) {
    if (errorCount < 3) {
      console.error(`[VU ${exec.vu.idInTest}] Fetch Large Article Failed. Status: ${res.status}, Error: ${res.error}, Body: ${res.body}`);
      errorCount++;
    }
  }
  
  check(res, {
    'is status 200': (r) => r.status === 200,
    'is large article': (r) => {
        try {
          return r.json('size_type') === 'large';
        } catch (e) {
          return false;
        }
    }
  });
}
