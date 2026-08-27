import http from 'k6/http';
import { check } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const baseURL = __ENV.SUB2API_TARGET_URL || 'http://127.0.0.1:8080';
const apiKey = __ENV.SUB2API_API_KEY || 'sk-local-replay';
const runID = __ENV.X_TEST_RUN_ID || `m111-${__VU}-${Date.now()}`;
const model = __ENV.SUB2API_MODEL || 'glm-5.2';
export const gateway_ttft_ms = new Trend('gateway_ttft_ms', true);
export const admission_wait_ms = new Trend('admission_wait_ms', true);
export const auth_decision_ms = new Trend('auth_decision_ms', true);
export const upstreamErrors = new Counter('upstream_errors');

export const options = {
  scenarios: {
    ramp_200_rps: { executor: 'ramping-arrival-rate', startRate: 0, timeUnit: '1s', preAllocatedVUs: 100, maxVUs: 500, stages: [{ target: 200, duration: '2m' }, { target: 200, duration: '5m' }, { target: 0, duration: '1m' }] },
    sustained_200_users: { executor: 'constant-vus', vus: 200, duration: __ENV.SOAK_DURATION || '10m', startTime: '1m' },
    burst_400_rps: { executor: 'ramping-arrival-rate', startRate: 0, timeUnit: '1s', preAllocatedVUs: 200, maxVUs: 800, stages: [{ target: 400, duration: '30s' }, { target: 400, duration: '1m' }, { target: 0, duration: '30s' }], startTime: '8m' },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    gateway_ttft_ms: ['p(95)<10000', 'p(99)<15000'],
    admission_wait_ms: ['p(99)<2000'],
    auth_decision_ms: ['p(95)<10'],
  },
};

export default function () {
  const mutation = __ITER % 10 === 0;
  const headers = { Authorization: `Bearer ${apiKey}`, 'Content-Type': 'application/json', 'X-Test-Run-Id': runID, 'X-Request-Id': `${runID}-${__VU}-${__ITER}` };
  if (mutation) headers['Idempotency-Key'] = `${runID}-mutation-${__VU}-${__ITER}`;
  const started = Date.now();
  const res = http.post(`${baseURL}/v1/chat/completions`, JSON.stringify({ model, messages: [{ role: 'user', content: 'health check' }], stream: false }), { headers, tags: { test_run_id: runID } });
  const elapsed = Date.now() - started;
  gateway_ttft_ms.add(elapsed);
  admission_wait_ms.add(Number(res.headers['X-Admission-Wait-Ms'] || 0));
  auth_decision_ms.add(Number(res.headers['X-Auth-Decision-Ms'] || 0));
  if (res.status >= 500) upstreamErrors.add(1);
  check(res, { 'gateway response is not 5xx': (r) => r.status < 500 });
}
