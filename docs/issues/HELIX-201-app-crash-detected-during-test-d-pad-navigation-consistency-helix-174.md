---
id: HELIX-201
severity: critical
category: functional
platform: androidtv
screen: Layout
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# App crash detected during test: D-pad Navigation Consistency (HELIX-174)

Stack trace: 04-04 10:25:32.973  4238 20385 E AndroidRuntime: FATAL EXCEPTION: Thread-70
04-04 10:25:32.973  4238 20385 E AndroidRuntime: java.lang.NullPointerException: Attempt to invoke virtual method 'int java.lang.String.length()' on a null object reference

Log entries: [04-04 10:25:32.973  4238 20385 E AndroidRuntime: FATAL EXCEPTION: Thread-70 04-04 10:25:32.973  4238 20385 E AndroidRuntime: java.lang.NullPointerException: Attempt to invoke virtual method 'int java.lang.String.length()' on a null object reference]

## Related Issues

- HELIX-155: App crash detected during test: Login Form Test
- HELIX-156: App crash detected during test: Register Form Test
- HELIX-157: App crash detected during test: Layout Test
- HELIX-158: App crash detected during test: Navigation Test
- HELIX-159: App crash detected during test: Entity Detail Test
- HELIX-160: App crash detected during test: Security Test
- HELIX-161: App crash detected during test: Known Issue Test - Insecure Password Input
- HELIX-162: App crash detected during test: Assets API Test
- HELIX-163: App crash detected during test: Browse API Test
- HELIX-164: App crash detected during test: Collections API Test
- HELIX-165: App crash detected during test: Challenges API Test
- HELIX-166: App crash detected during test: Known Issue Test - Lack of Feedback on Form Submission
- HELIX-167: App crash detected during test: Error Handling Test
- HELIX-168: App crash detected during test: Performance Test
- HELIX-169: App crash detected during test: Known Issue Test - Inconsistent Label Alignment
- HELIX-179: App crash detected during test: Insecure Password Input Test
- HELIX-180: App crash detected during test: Memory Leak Test
- HELIX-181: App crash detected during test: Lack of Feedback on Form Submission Test
- HELIX-182: App crash detected during test: Inconsistent Label Alignment Test
- HELIX-184: App crash detected during test: Successful User Login (LoginForm)
- HELIX-185: App crash detected during test: User Registration with Valid Data (RegisterForm)
- HELIX-186: App crash detected during test: Navigation Menu Accessibility (HELIX-055, HELIX-113, HELIX-125)
- HELIX-187: App crash detected during test: Search Functionality with Valid Query (HELIX-035, HELIX-041, HELIX-062, HELIX-114, HELIX-122, HELIX-153)
- HELIX-188: App crash detected during test: Entity Detail Page Load and Display
- HELIX-189: App crash detected during test: Memory Leak Test (HELIX-178, HELIX-180)
- HELIX-190: App crash detected during test: Unreachable API Handling (HELIX-037, HELIX-038, HELIX-167)
- HELIX-191: App crash detected during test: Accessibility of Keyboard for Users with Disabilities (HELIX-177)
- HELIX-192: App crash detected during test: App Crash During Navigation (HELIX-158)
- HELIX-193: App crash detected during test: Incorrect Password Login Attempt
- HELIX-194: App crash detected during test: Password Field Display (HELIX-093)
- HELIX-195: App crash detected during test: Password Field 'Show/Hide' Option (HELIX-006)
- HELIX-196: App crash detected during test: Search Functionality with No Results (HELIX-031, HELIX-138, HELIX-153)
- HELIX-197: App crash detected during test: API Endpoint: GET /api/v1/media/:id - Valid Request
- HELIX-198: App crash detected during test: API Endpoint: GET /api/v1/media/:id - Invalid ID
- HELIX-199: App crash detected during test: API Endpoint: POST /api/v1/auth/login - Invalid Credentials
- HELIX-200: App crash detected during test: Keyboard Layout Optimization for Search (HELIX-059, HELIX-140, HELIX-176)


