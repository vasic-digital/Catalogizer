# Catalogizer Project Constitution
## Core Development & QA Principles

### Article I: Real-Time Log Monitoring (MANDATORY)

**§1.1 All QA Sessions Must Monitor Logs in Real-Time**

During ANY QA session execution (HelixQA or manual testing), real-time log monitoring is **MANDATORY** and **NON-NEGOTIABLE** for:

- **Android/Android TV Apps**: `adb logcat` must be actively monitored
- **Web Applications**: Browser console logs must be captured and monitored
- **Desktop Applications**: Application logs and system logs must be monitored
- **Backend Services**: Service logs, error logs, and access logs must be monitored
- **All Services**: Any running service or daemon must have its logs monitored

**§1.2 Purpose**

Real-time log monitoring ensures immediate detection of:
- Application Not Responding (ANR) errors
- Fatal crashes and exceptions
- Memory leaks and OOM errors
- Network timeouts and connection failures
- Security violations and unauthorized access attempts
- Performance degradation indicators

**§1.3 Implementation Requirements**

1. **HelixQA must capture and stream logs** for all tested platforms simultaneously
2. **Log analysis must happen in real-time**, not post-session
3. **ANR/Crash detection must trigger immediate alerts** and session pausing
4. **All log outputs must be saved** to the session directory for post-analysis
5. **Log monitoring is NOT optional** - no QA session is valid without it

**§1.4 Violation Consequences**

Any QA session conducted without real-time log monitoring is **INVALID** and must be repeated.

---

### Article II: Video Recording (MANDATORY)

**§2.1 All QA Sessions Must Record Video**

Video recording is **MANDATORY** for all UI/UX QA sessions. Screen captures alone are insufficient.

---

### Article III: Crash Fix Protocol

**§3.1 Immediate Action Required**

When ANRs or crashes are detected:
1. **STOP** the QA session immediately
2. **CAPTURE** all relevant logs and stack traces
3. **ANALYZE** the root cause
4. **FIX** the underlying issue in the application code
5. **VERIFY** the fix with a focused regression test
6. **RESUME** full QA only after verification passes

---

*Last Updated: 2026-04-08*
*Enforced by: Project Lead*
