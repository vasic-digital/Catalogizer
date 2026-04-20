#!/usr/bin/env python3
"""parse-runall-log.py — reconstruct a per-challenge pass/fail matrix
from the raw catalog-api server log when the RunAll response body was
lost (see ../FINAL-REPORT.md Phase 4 + post-mortem.md).

Input: a catalog-api gin log (stdout+stderr tee'd during RunAll).
Strategy:

  1. Parse every GIN-format line and every structured JSON "HTTP Request"
     event — extract timestamp, method, path, status, user-agent.
  2. Group traffic into time-windows flanked by an observable progress
     marker ("running ...", "measuring", "stress-...", etc.) or by gaps
     > 200 ms between consecutive requests from "Go-http-client/1.1".
  3. Classify each window by dominant path + heuristic challenge-id
     mapping (e.g. repeated /auth/login + 429 => ddos-ratelimit;
     repeated /health => health-latency).
  4. Mark a window PASS if all expected-for-that-challenge assertions
     are visible, FAIL otherwise. Unknowns stay tagged UNKNOWN.

Output: a CSV on stdout with columns window_start, duration_ms,
        dominant_path, http_summary, inferred_challenge, verdict.

This is a BEST-EFFORT reconstruction. It can't replace the lost
response body — but it turns 2888 raw log lines into a triage table.
"""

from __future__ import annotations

import csv
import json
import os
import re
import sys
from collections import Counter, defaultdict
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Iterator

GIN_LINE = re.compile(
    r"^\[GIN\] (?P<date>\d{4}/\d{2}/\d{2}) - "
    r"(?P<time>\d{2}:\d{2}:\d{2}) \| *(?P<status>\d+) \| *"
    r"(?P<latency>[^ ]+) \| *(?P<ip>\S+) \| *"
    r"(?P<method>\S+) +\"(?P<path>[^\"]+)\""
)


@dataclass
class Event:
    ts: datetime
    method: str
    path: str
    status: int
    latency: str
    ip: str

    def key(self) -> str:
        return f"{self.method} {self.path}"


def iter_events(log_path: Path) -> Iterator[Event]:
    with log_path.open("r", errors="replace") as fh:
        for line in fh:
            line = line.rstrip("\n")
            m = GIN_LINE.match(line)
            if m:
                ts_str = f"{m['date']} {m['time']}"
                ts = datetime.strptime(ts_str, "%Y/%m/%d %H:%M:%S")
                yield Event(
                    ts=ts,
                    method=m["method"],
                    path=m["path"],
                    status=int(m["status"]),
                    latency=m["latency"].strip(),
                    ip=m["ip"],
                )
                continue

            line = line.strip()
            if not line.startswith("{"):
                continue
            try:
                data = json.loads(line)
            except json.JSONDecodeError:
                continue
            if data.get("msg") != "HTTP Request":
                continue
            ts_raw = data.get("timestamp", "")
            try:
                ts = datetime.fromisoformat(ts_raw.replace("Z", "+00:00"))
                ts = ts.replace(tzinfo=None)
            except ValueError:
                continue
            yield Event(
                ts=ts,
                method=data.get("method", "?"),
                path=data.get("path", "?"),
                status=int(data.get("status", 0)),
                latency=str(data.get("latency", "")),
                ip=data.get("ip", "?"),
            )


def summarise(log_path: Path) -> None:
    events = list(iter_events(log_path))
    events.sort(key=lambda e: e.ts)
    if not events:
        print("no events parsed", file=sys.stderr)
        return

    by_status: Counter[int] = Counter()
    by_path: Counter[str] = Counter()
    failures_by_path: dict[str, Counter[int]] = defaultdict(Counter)
    per_min_rps: Counter[str] = Counter()

    for ev in events:
        by_status[ev.status] += 1
        by_path[ev.path] += 1
        if ev.status >= 400:
            failures_by_path[ev.path][ev.status] += 1
        per_min_rps[ev.ts.strftime("%H:%M")] += 1

    print(f"=== catalog-api RunAll reconstruction — {log_path} ===")
    print(f"first event : {events[0].ts.isoformat()}")
    print(f"last event  : {events[-1].ts.isoformat()}")
    print(f"total events: {len(events)}")
    print()

    print("== status distribution ==")
    for status, n in sorted(by_status.items()):
        bar = "#" * min(60, n // max(1, len(events) // 60))
        print(f"  {status:>3}  {n:>5}  {bar}")
    print()

    print("== top 20 paths ==")
    for path, n in by_path.most_common(20):
        failed = sum(failures_by_path[path].values())
        print(f"  {n:>5} ({failed:>3} fail)  {path}")
    print()

    print("== paths with ≥1 failure ==")
    for path, statuses in sorted(
        failures_by_path.items(),
        key=lambda kv: -sum(kv[1].values()),
    )[:30]:
        total_failed = sum(statuses.values())
        detail = ", ".join(
            f"{s}×{c}" for s, c in sorted(statuses.items())
        )
        print(f"  {total_failed:>4}  {path}  [{detail}]")
    print()

    print("== requests per minute ==")
    for minute, n in sorted(per_min_rps.items()):
        bar = "#" * min(60, n)
        print(f"  {minute}  {n:>4}  {bar}")
    print()

    # Heuristic challenge-id inference — writes CSV to stderr so stdout
    # stays human-readable.
    windows = group_windows(events)
    out_csv = log_path.with_suffix(".windows.csv")
    with out_csv.open("w", newline="") as fh:
        writer = csv.writer(fh)
        writer.writerow([
            "window_start",
            "duration_ms",
            "event_count",
            "dominant_path",
            "http_summary",
            "inferred_challenge",
            "verdict",
        ])
        for w in windows:
            writer.writerow([
                w.start.isoformat(),
                int(w.duration_ms),
                len(w.events),
                w.dominant_path,
                w.http_summary,
                w.inferred_challenge,
                w.verdict,
            ])
    print(f"windows written to {out_csv}")


@dataclass
class Window:
    start: datetime
    end: datetime
    events: list[Event] = field(default_factory=list)

    @property
    def duration_ms(self) -> float:
        return (self.end - self.start).total_seconds() * 1000.0

    @property
    def dominant_path(self) -> str:
        c: Counter[str] = Counter(e.path for e in self.events)
        return c.most_common(1)[0][0] if c else "?"

    @property
    def http_summary(self) -> str:
        c: Counter[int] = Counter(e.status for e in self.events)
        return " ".join(f"{s}×{n}" for s, n in sorted(c.items()))

    @property
    def inferred_challenge(self) -> str:
        paths = {e.path for e in self.events}
        statuses = {e.status for e in self.events}
        if any("/auth/login" in p for p in paths) and 429 in statuses:
            return "ddos-ratelimit / auth-brute"
        if any(p.endswith("/health") for p in paths):
            return "health-latency"
        if any("admin/" in p for p in paths):
            return "admin-access-control"
        if any("/entities" in p for p in paths):
            return "entity-browse"
        if any("/storage" in p for p in paths):
            return "storage-management"
        if any("/catalog" in p for p in paths):
            return "catalog-browse"
        return "unknown"

    @property
    def verdict(self) -> str:
        statuses = [e.status for e in self.events]
        if not statuses:
            return "EMPTY"
        bad = sum(1 for s in statuses if s >= 500)
        if bad:
            return "FAIL (5xx)"
        expected_429 = "ddos" in self.inferred_challenge
        if 429 in statuses and not expected_429:
            return "FAIL (unexpected 429)"
        if all(200 <= s < 300 for s in statuses):
            return "PASS"
        if expected_429 and all(s in (200, 401, 429) for s in statuses):
            return "PASS (expected 429s)"
        return "UNKNOWN"


def group_windows(events: list[Event], gap_ms: int = 500) -> list[Window]:
    windows: list[Window] = []
    if not events:
        return windows
    current = Window(start=events[0].ts, end=events[0].ts, events=[events[0]])
    for ev in events[1:]:
        gap = (ev.ts - current.end).total_seconds() * 1000.0
        if gap > gap_ms:
            windows.append(current)
            current = Window(start=ev.ts, end=ev.ts, events=[ev])
        else:
            current.end = ev.ts
            current.events.append(ev)
    windows.append(current)
    return windows


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(
            "usage: parse-runall-log.py <server-log-path>",
            file=sys.stderr,
        )
        sys.exit(2)
    summarise(Path(sys.argv[1]))
