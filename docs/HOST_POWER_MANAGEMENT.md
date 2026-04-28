# Host Power Management — Hard Ban (CONST-033)

## Why this exists

On 2026-04-26 18:23:43 the host running mission-critical parallel CLI
agents and container workloads was auto-suspended mid-session. This
killed the running HelixAgent binary, all 41 dependent services, every
SSH session, and every active CLI agent on the box. journalctl showed:

```
systemd-logind[1183]: The system will suspend now!
```

The user-level GNOME power settings were already correct
(`sleep-inactive-ac-type=nothing`). The trigger was the **GDM greeter
session at the local console**, which has its own power policy and
does not count SSH activity. Earlier, on multiple occasions, the
user@1000.service had been SIGKILLed by systemd because heavy memory
pressure prevented gnome-shell from responding to GDM/Wayland watchdog
within `TimeoutStopSec` — perceived by the user as "the system fully
logged me out." Together these two failure modes have caused repeated
loss of in-flight agent work.

## The rule

**No project shipped from this workspace may invoke a host-level
power-state transition.** Forbidden invocations include — but are not
limited to:

| Layer | Forbidden invocations |
|-------|------------------------|
| systemd CLI | `systemctl suspend`, `systemctl hibernate`, `systemctl hybrid-sleep`, `systemctl suspend-then-hibernate`, `systemctl poweroff`, `systemctl halt`, `systemctl reboot`, `systemctl kexec` |
| logind CLI  | `loginctl suspend`, `loginctl hibernate`, `loginctl hybrid-sleep`, `loginctl suspend-then-hibernate`, `loginctl poweroff`, `loginctl halt`, `loginctl reboot` |
| Legacy CLI  | `pm-suspend`, `pm-hibernate`, `pm-suspend-hybrid`, `shutdown -h/-r/-P/-H/now`, bare `reboot`/`poweroff`/`halt` |
| DBus | `org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}`, `org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}` (via `dbus-send`, `busctl`, or any language binding) |
| gsettings | `gsettings set ... sleep-inactive-{ac,battery}-type` to anything other than `'nothing'` or `'blank'` |

## Defence in depth

Three layers, in order of strength:

1. **Host-level masking (manual prereq, sudo required, run once):**
   `scripts/host-power-management/install-host-suspend-guard.sh`
   masks `sleep.target`, `suspend.target`, `hibernate.target`,
   `hybrid-sleep.target`, writes `/etc/systemd/sleep.conf.d/00-no-suspend.conf`
   with `AllowSuspend=no`, and writes
   `/etc/systemd/logind.conf.d/00-no-idle-suspend.conf` with
   `IdleAction=ignore` and `HandleLidSwitch=ignore`. After this, no
   user / session / DE / greeter / cron job can suspend the host.

2. **User-session bootstrap (no sudo):**
   `scripts/host-power-management/user_session_no_suspend_bootstrap.sh`
   runs `gsettings`, `xset -dpms`, and (opt-in via
   `HOST_POWER_MANAGEMENT_SESSION_INHIBIT=1`) `systemd-inhibit` to
   protect the current GUI/CLI session of the invoking user. Idempotent.
   Safe to source from `start.sh` / `setup.sh` / `bootstrap.sh`.

3. **Source-tree static gate:**
   `scripts/host-power-management/check-no-suspend-calls.sh` walks the
   tree and exits non-zero on any forbidden invocation.
   `challenges/scripts/no_suspend_calls_challenge.sh` wraps it as a
   challenge that runs in CI / `run_all_challenges.sh`.
   `challenges/scripts/host_no_auto_suspend_challenge.sh` asserts the
   running host's state matches the layer-1 masking.

## Verification

```bash
# Layer 1: host state
bash challenges/scripts/host_no_auto_suspend_challenge.sh
# Expected: 4 PASS

# Layer 3: source tree
bash challenges/scripts/no_suspend_calls_challenge.sh
# Expected: PASS, no forbidden calls
```

## If you genuinely need a power-state transition

You don't, in this workspace. If a future legitimate use case arises
(e.g. a container's *internal* init script that suspends a *guest VM*,
not the host), add the specific file path to `EXCLUDE_PATHS` at the
top of `check-no-suspend-calls.sh` with a comment explaining the
non-host context.

## Container runtimes as a session-loss vector (§12 Incident #2 — 2026-04-28)

CONST-033 bans **direct** power-management calls. But a host can still
lose its user session — same observable outcome — through **indirect**
container-runtime pressure on `user@<uid>.service`. This section
documents what we've observed and how to defend against it.

### What happened

On 2026-04-28 18:36:35 MSK the user's `user@1000.service` was SIGKILLed
(`status=9/KILL`) **without** a kernel OOM kill (`systemd-oomd`
inactive, `MemoryMax=infinity`). Cascade killed `claude`, `tmux`, an
in-flight build, and 20+ npm-spawned MCP server processes. From the
operator's perspective this was indistinguishable from a
suspend/hibernate/logout — same lost work, same broken Claude session,
same need to resume from scratch — but it was **not** triggered by any
power-management call.

Six minutes before the kill, the journal contained recurring conmon
warnings of the form:

```
conmon[…]: Failed to open cgroups file: /sys/fs/cgroup/memory.events
```

Six minutes is suspicious. The forensic conclusion in the parent
ATMOSphere repo's `docs/guides/ATMOSPHERE_CONSTITUTION.md` §12
Incident #2 is that **cumulative cgroup pressure from many concurrent
container workloads, plus an external watchdog observing the conmon
warnings, can escalate to user-slice termination** even when no single
container exceeded its `mem_limit`. Per-container limits don't bound
the user-slice; they bound the container.

### Mitigations (mandatory in this workspace)

1. **Per-container memory limit, always.** Every `podman run` /
   `docker run` / compose service MUST declare `--memory=<bytes>` /
   `mem_limit:`. Unbounded containers are forbidden — a single
   runaway can drown the user-slice without tripping any
   per-container limit.
2. **Σ mem_limit ≤ physical RAM − user-session overhead.** This is
   a manual operator check before launching big workloads. The
   host's `user@1000.service` itself wants ~3–4 GB headroom for the
   shell, IDE, browser, and Claude Code. Sum the planned limits and
   subtract.
3. **No `--ulimit memlock=infinity`** unless absolutely required
   (locks pages out of swap reach).
4. **`OOMPolicy=stop` in systemd container units.** Avoid the default
   `restart-on-OOM` loop — it pumps memory pressure indefinitely.
5. **Exponential-backoff restart, never immediate.** A failing
   container in a tight restart loop is the second-most-common
   trigger of host instability after unbounded memory.
6. **Clean-slate destroy after any §12 incident.** After a host
   crash or session-loss event, **always** `podman ps -a → rm -f`
   and `podman volume prune` for the affected stack. Stale lock
   files (`/run/containers/storage/...`, `/var/lib/containers/...`,
   `~/.local/share/containers/storage/locks`) keep producing
   inscrutable failures otherwise.
7. **Monitor `conmon` warnings.** If
   `journalctl -k --since "1 hour ago" | grep -c "Failed to open cgroups file"`
   is non-zero, **fix the offending workload first**. Do not stack
   new heavy work on a host already in distress.
8. **MCP-server multiplier warning.** 20+ npm-spawned MCP server
   processes are a known memory multiplier. Stop non-essential MCPs
   before heavy build/test/QA work. The pattern "my Claude session
   has 25 MCP servers running" is the highest-risk configuration
   this project has encountered.

### Forensic checklist when the symptom recurs

1. `last -x` and `last reboot` — confirm shutdown vs reboot vs
   session kill.
2. `journalctl --user -b -1 -p err` — last user-session errors before
   logout.
3. `journalctl -b -1 | grep -iE "oom-kill|systemd-oomd|user@1000.*KILL"` —
   was there an OOM cascade? An external SIGKILL?
4. `journalctl -b -1 | grep -c "Failed to open cgroups file"` — how
   many conmon warnings preceded the event, and at what timestamp?
5. `journalctl -b -1 | grep -E "podman|docker|conmon" | tail -200` —
   container-runtime activity in the lead-up.
6. `bash challenges/scripts/no_suspend_calls_challenge.sh` and
   `bash challenges/scripts/host_no_auto_suspend_challenge.sh` —
   confirm CONST-033 is still in force.

If conmon warnings are non-zero **and** the SIGKILL was not
OOM-killed, it's a §12-class event. Document under
`docs/incidents/<YYYY-MM-DD>-…md`, clean-slate the containers, and
review the active workload's cumulative memory budget before
bringing the stack back up.

### Cross-references

- Containers submodule: `Containers/CLAUDE.md` "MANDATORY HOST-SESSION
  SAFETY (Constitution §12)" + "MANDATORY §12 HOST-SESSION SAFETY —
  INCIDENT #2 ANCHOR (2026-04-28)".
- Parent ATMOSphere project (canonical authority):
  `docs/guides/ATMOSPHERE_CONSTITUTION.md` §12.
- Project Constitution: Article X (no-sudo / no-root) + CONST-033
  (host power management) + Article XI §11.5 (real-system test
  requirements that need disciplined per-container budgets).
