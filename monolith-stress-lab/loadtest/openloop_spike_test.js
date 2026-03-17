import http from 'k6/http';
import { check } from 'k6';

const qps = __ENV.QPS ? parseInt(__ENV.QPS) : 100;
const duration = __ENV.DURATION || '10s';
const baseUrl = __ENV.BASE_URL || 'http://app:8080';
const maxOpen = __ENV.DB_MAX_OPEN ? parseInt(__ENV.DB_MAX_OPEN) : null;
const cost = __ENV.COST ? parseInt(__ENV.COST) : null;

// The constant-arrival-rate executor tries to start `rate` iterations every `timeUnit`
// regardless of how long each iteration takes. If the server is slow, K6 will spin up
// more VUs up to `preAllocatedVUs` (and optionally `maxVUs`) to maintain the pace.
export const options = {
  scenarios: {
    // 80% Read Heavy
    openloop_read: {
      executor: 'constant-arrival-rate',
      rate: Math.floor(qps * 0.8),
      timeUnit: '1s',
      duration: duration,
      preAllocatedVUs: Math.min(20000, Math.floor(qps * 0.8)), 
      maxVUs: 48000,               // 暴力流：允许无限派兵直到 48000 个并发读连接
      gracefulStop: '3s',
      exec: 'read_articles',
    },
    // 10% Write Heavy
    openloop_write: {
      executor: 'constant-arrival-rate',
      rate: Math.max(1, Math.floor(qps * 0.1)),
      timeUnit: '1s',
      duration: duration,
      preAllocatedVUs: Math.min(2000, Math.floor(qps * 0.1)),
      maxVUs: 6000,                // 暴力流：允许无限派兵直到 6000 个并发写连接
      gracefulStop: '3s',
      exec: 'post_comments',
    },
    // 10% CPU Heavy Login
    openloop_login: {
      executor: 'constant-arrival-rate',
      rate: Math.max(1, Math.floor(qps * 0.1)),
      timeUnit: '1s',
      duration: duration,
      preAllocatedVUs: Math.min(2000, Math.floor(qps * 0.1)),
      maxVUs: 6000,                // 暴力流：允许无限派兵直到 6000 个并发计算连接
      gracefulStop: '3s',
      exec: 'login_users',
    },
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export function setup() {
  console.log(`[*] Open-Loop Spike Test: Target QPS=${qps} (80% Read, 10% Write, 10% Login)`);
  if (maxOpen) {
    const res = http.put(`${baseUrl}/config/db-pool`, JSON.stringify({ max_open: maxOpen, max_idle: maxOpen }), { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 200) console.log(`[+] Auto-tuning DB Pool to ${maxOpen} SUCCESS`);
  }
  if (cost) {
    const res = http.put(`${baseUrl}/config/bcrypt-cost`, JSON.stringify({ cost: cost }), { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 200) console.log(`[+] Auto-tuning Bcrypt Cost to ${cost} SUCCESS`);
  }
}

export function read_articles() {
  const id = Math.floor(Math.random() * 100) + 101; 
  const res = http.get(`${baseUrl}/articles/${id}`, { timeout: '500ms' });
  check(res, { 'Read Article status 200': (r) => r.status === 200 });
}

export function post_comments() {
  const userId = Math.floor(Math.random() * 10000) + 1;
  const articleId = Math.floor(Math.random() * 200) + 1;
  const payload = JSON.stringify({ user_id: userId, article_id: articleId, content: `Spike comment ${Date.now()}` });
  const res = http.post(`${baseUrl}/comments`, payload, { headers: { 'Content-Type': 'application/json' }, timeout: '500ms' });
  check(res, { 'Post Comment status 201': (r) => r.status === 201 });
}

export function login_users() {
  const userIndex = Math.floor(Math.random() * 10000) + 1;
  const payload = JSON.stringify({ username: `user${userIndex}`, password: '123456' });
  const res = http.post(`${baseUrl}/login`, payload, { headers: { 'Content-Type': 'application/json' }, timeout: '500ms' });
  check(res, { 'Login status 200': (r) => r.status === 200 });
}
