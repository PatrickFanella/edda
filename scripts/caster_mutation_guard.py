"""Durable serial mutation pacing; timing is mitigation, never retry authority."""
from datetime import datetime
from email.utils import parsedate_to_datetime
import json
import re
import subprocess
import time

SAFE_HEADERS = {'date', 'x-github-request-id', 'x-ratelimit-limit',
                'x-ratelimit-remaining', 'x-ratelimit-used', 'x-ratelimit-reset',
                'x-ratelimit-resource', 'retry-after'}
NOT_BEFORE = datetime.fromisoformat('2026-11-17T00:04:50+00:00').timestamp()
MIN_INTERVAL = 7.5
# ROOT state-only reopen PATCHes; these do not prove createIssue recovery.
EXTERNAL_STARTS = [timestamp for timestamp in (
    datetime.fromisoformat('2026-11-17T00:00:02+00:00').timestamp(),
    datetime.fromisoformat('2026-11-17T00:00:04+00:00').timestamp(),
    datetime.fromisoformat('2026-11-17T00:00:06+00:00').timestamp())]


def timestamp(value):
    return datetime.fromisoformat(value.replace('Z', '+00:00')).timestamp()


def parse_http(raw):
    head, separator, body = raw.replace('\r\n', '\n').partition('\n\n')
    assert separator and re.match(r'^HTTP/\S+ \d{3}', head), 'HTTP response envelope missing'
    headers = {}
    for line in head.splitlines()[1:]:
        key, colon, value = line.partition(':')
        if colon and key.lower() in SAFE_HEADERS:
            headers[key.lower()] = value.strip()
    # Slice original bytes-as-text to preserve CRLF in the GraphQL response body.
    body = raw.split('\r\n\r\n', 1)[1] if '\r\n\r\n' in raw else raw.split('\n\n', 1)[1]
    return dict(status=int(head.splitlines()[0].split()[1]), headers=headers, body=body)


def http_graphql(query, variables):
    result = subprocess.run(['gh', 'api', 'graphql', '--include', '--input', '-'],
                            input=json.dumps(dict(query=query, variables=variables)),
                            text=True, capture_output=True, timeout=90)
    response = parse_http(result.stdout)
    response['exit_code'] = result.returncode
    return response


class MutationGuard:
    def __init__(self, journal, clock=time.time, sleep=time.sleep):
        self.journal, self.clock, self.sleep = journal, clock, sleep

    def deadline(self, extra_responses=()):
        deadline = NOT_BEFORE
        responses = list(extra_responses)
        starts = list(EXTERNAL_STARTS)
        for entry in self.journal['operations']:
            if 'started_at' not in entry or entry.get('status') == 'NOT_DISPATCHED':
                continue
            start = timestamp(entry['started_at'])
            starts.append(start)
            deadline = max(deadline, start + MIN_INTERVAL)
            if entry.get('status') in ('INTENT', 'RECONCILE_REQUIRED'):
                deadline = max(deadline, timestamp(entry.get('finished_at', entry['started_at'])) + 3600)
            response = entry.get('sanitized_http_response')
            if response:
                responses.append((response, timestamp(entry.get('finished_at', entry['started_at']))))
        if starts:
            deadline = max(deadline, max(starts) + MIN_INTERVAL)
        if len(starts) >= 480:
            deadline = max(deadline, sorted(starts)[-480] + 3600.001)
        for response, observed in responses:
            headers = response['headers']
            if 'retry-after' in headers:
                value = headers['retry-after']
                until = observed + float(value) if value.isdecimal() else parsedate_to_datetime(value).timestamp()
                deadline = max(deadline, until)
            if headers.get('x-ratelimit-remaining') == '0':
                assert 'x-ratelimit-reset' in headers, 'primary reset missing; writes held'
                deadline = max(deadline, float(headers['x-ratelimit-reset']) + 1)
        return deadline

    def wait(self, extra_responses=()):
        # Never block an agent for an hour. Long restrictions require a later run.
        delay = self.deadline(extra_responses) - self.clock()
        assert delay <= 60, 'mutation cooldown active; resume after deadline'
        while delay > 0:
            self.sleep(delay)
            delay = self.deadline(extra_responses) - self.clock()
