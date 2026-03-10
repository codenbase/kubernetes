import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';

// Read configuration from environment or use defaults
const maxOpen = __ENV.DB_MAX_OPEN ? parseInt(__ENV.DB_MAX_OPEN) : null;
const cost = __ENV.COST ? parseInt(__ENV.COST) : null;
const baseUrl = __ENV.BASE_URL || 'http://app:8080';

export function setup() {
  console.log('[*] Mixed Workload Test Setup Started');
  
  if (maxOpen) {
    const payload = JSON.stringify({ max_open: maxOpen, max_idle: maxOpen });
    const res = http.put(`${baseUrl}/config/db-pool`, payload, { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 200) console.log(`[+] Auto-tuning DB Pool (MaxOpen) to ${maxOpen} SUCCESS`);
    else console.error(`[-] Failed to tune DB Pool: ${res.status} ${res.body}`);
  }
  
  if (cost) {
    const payload = JSON.stringify({ cost: cost });
    const res = http.put(`${baseUrl}/config/bcrypt-cost`, payload, { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 200) console.log(`[+] Auto-tuning Bcrypt Cost to ${cost} SUCCESS`);
    else console.error(`[-] Failed to tune Bcrypt Cost: ${res.status} ${res.body}`);
  }
}

export const options = {
  scenarios: {
    // 80% Read Heavy (Browsing Articles)
    read_heavy: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 80 },  // Phase 1: Climb
        { duration: '1m', target: 240 },   // Phase 2: Spike
        { duration: '1m', target: 480 },   // Phase 3: Final Push
        { duration: '30s', target: 0 }     // Cooldown
      ],
      exec: 'read_articles',
    },
    // 10% Write Heavy (Posting Comments)
    write_heavy: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 30 },
        { duration: '1m', target: 60 },
        { duration: '30s', target: 0 }
      ],
      exec: 'post_comments',
    },
    // 10% CPU Heavy (Logins)
    cpu_heavy: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 30 },
        { duration: '1m', target: 60 },
        { duration: '30s', target: 0 }
      ],
      exec: 'login_users',
    },
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export function read_articles() {
  // Mostly read large articles (simulating realistic rich content fetching)
  const id = Math.floor(Math.random() * 100) + 101; 
  const res = http.get(`${baseUrl}/articles/${id}`);
  check(res, { 'Read Article status 200': (r) => r.status === 200 });
}

export function post_comments() {
  const userId = Math.floor(Math.random() * 10000) + 1;
  const articleId = Math.floor(Math.random() * 200) + 1;
  const payload = JSON.stringify({
    user_id: userId,
    article_id: articleId,
    content: `Mixed workload comment generated at ${Date.now()}`
  });
  const res = http.post(`${baseUrl}/comments`, payload, { headers: { 'Content-Type': 'application/json' } });
  check(res, { 'Post Comment status 201': (r) => r.status === 201 });
}

export function login_users() {
  const userIndex = Math.floor(Math.random() * 10000) + 1;
  const payload = JSON.stringify({ username: `user${userIndex}`, password: '123456' });
  const res = http.post(`${baseUrl}/login`, payload, { headers: { 'Content-Type': 'application/json' } });
  check(res, { 'Login status 200': (r) => r.status === 200 });
}
