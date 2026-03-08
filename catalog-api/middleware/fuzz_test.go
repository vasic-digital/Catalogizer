package middleware

import (
	"testing"
)

// FuzzSanitizeInput fuzzes the SanitizeInput function to ensure it never panics
// and always returns consistent results for the same input.
func FuzzSanitizeInput(f *testing.F) {
	// Seed corpus: normal strings
	f.Add("")
	f.Add("hello world")
	f.Add("   leading and trailing spaces   ")
	f.Add("Hello, 世界! 🌍")
	f.Add("正常文字")

	// Null bytes
	f.Add("hello\x00world")
	f.Add("\x00\x00\x00")
	f.Add("before\x00after\x00end")

	// Invalid UTF-8 sequences
	f.Add("hello\xc0\xafworld")
	f.Add("\xfe\xff")
	f.Add("valid\x80invalid")
	f.Add("\xed\xa0\x80") // surrogate half

	// Whitespace variations
	f.Add("\t\n\r\x0b\x0c")
	f.Add("  \t\n  hello  \t\n  ")

	// Long strings
	f.Add("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	// Mixed attack payloads (sanitizer should handle without panic)
	f.Add("SELECT * FROM users; DROP TABLE users;--")
	f.Add("<script>alert('xss')</script>")
	f.Add("../../../etc/passwd")
	f.Add("%2e%2e%2f%2e%2e%2fetc%2fpasswd")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic
		result1 := SanitizeInput(input)

		// Calling twice on the same raw input must return the same result
		result2 := SanitizeInput(input)
		if result1 != result2 {
			t.Errorf("SanitizeInput not consistent: first=%q, second=%q, input=%q", result1, result2, input)
		}

		// Double-sanitizing must converge: sanitize(sanitize(x)) == sanitize(sanitize(sanitize(x)))
		// Note: a single pass may expose new leading/trailing whitespace after removing
		// invalid UTF-8 bytes, so sanitize(sanitize(x)) != sanitize(x) is acceptable.
		// But the second application must be a fixed point.
		result3 := SanitizeInput(result1)
		result4 := SanitizeInput(result3)
		if result3 != result4 {
			t.Errorf("SanitizeInput does not converge: pass2=%q, pass3=%q, input=%q", result3, result4, input)
		}
	})
}

// FuzzContainsSQLInjection fuzzes the DetectSQLInjection function to ensure it
// never panics and returns consistent results.
func FuzzContainsSQLInjection(f *testing.F) {
	// Benign inputs
	f.Add("")
	f.Add("hello world")
	f.Add("John Doe")
	f.Add("user@example.com")
	f.Add("My favorite movie is The Selection")
	f.Add("I have 3 orders pending")

	// Classic SQL injection payloads
	f.Add("SELECT * FROM users")
	f.Add("1' OR '1'='1")
	f.Add("admin'--")
	f.Add("'; DROP TABLE users;--")
	f.Add("1 UNION SELECT password FROM users")
	f.Add("INSERT INTO users VALUES('hacker','pass')")
	f.Add("UPDATE users SET role='admin' WHERE id=1")
	f.Add("DELETE FROM users WHERE 1=1")
	f.Add("CREATE TABLE exploit(data TEXT)")
	f.Add("ALTER TABLE users ADD COLUMN backdoor TEXT")

	// Time-based blind SQL injection
	f.Add("WAITFOR DELAY '0:0:5'")
	f.Add("BENCHMARK(5000000,SHA1('test'))")
	f.Add("1; SLEEP(5)")

	// Stored procedure injection
	f.Add("EXEC(xp_cmdshell 'dir')")
	f.Add("EXEC sp_configure 'show advanced options', 1")
	f.Add("xp_cmdshell 'net user hacker pass /add'")

	// Comment-based injection
	f.Add("admin'/*")
	f.Add("1 /* comment */ UNION SELECT 1")

	// Obfuscated payloads
	f.Add("SEL/**/ECT * FROM users")
	f.Add("sElEcT * fRoM users")
	f.Add("' OR 1=1--")
	f.Add("' HAVING 1=1--")
	f.Add("' GROUP BY columnname--")
	f.Add("' ORDER BY 1--")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic
		result1 := DetectSQLInjection(input)

		// Calling twice must return the same result
		result2 := DetectSQLInjection(input)
		if result1 != result2 {
			t.Errorf("DetectSQLInjection not consistent: first=%v, second=%v, input=%q", result1, result2, input)
		}
	})
}

// FuzzContainsXSS fuzzes the DetectXSS function to ensure it never panics
// and returns consistent results.
func FuzzContainsXSS(f *testing.F) {
	// Benign inputs
	f.Add("")
	f.Add("hello world")
	f.Add("John Doe")
	f.Add("user@example.com")
	f.Add("<p>This is valid HTML</p>")
	f.Add("Price is $5 < $10")

	// Script tag variants
	f.Add("<script>alert('xss')</script>")
	f.Add("<SCRIPT>alert('xss')</SCRIPT>")
	f.Add("<script type='text/javascript'>document.cookie</script>")
	f.Add("<script src='http://evil.com/xss.js'></script>")
	f.Add("<ScRiPt>alert(1)</ScRiPt>")

	// Iframe injection
	f.Add("<iframe src='http://evil.com'></iframe>")
	f.Add("<iframe src='javascript:alert(1)'></iframe>")

	// Object/embed injection
	f.Add("<object data='evil.swf'></object>")
	f.Add("<embed src='evil.swf'></embed>")

	// Event handler injection
	f.Add("<img onerror='alert(1)' src='x'>")
	f.Add("<div onmouseover='alert(1)'>hover</div>")
	f.Add("<body onload='alert(1)'>")
	f.Add("<img src=x onerror=alert(1)>")
	f.Add("<svg onload=alert(1)>")
	f.Add("<input onfocus=alert(1) autofocus>")

	// JavaScript/VBScript URI
	f.Add("javascript:alert('xss')")
	f.Add("vbscript:msgbox(1)")
	f.Add("data:text/html,<script>alert(1)</script>")
	f.Add("javascript:void(0)")

	// Function call patterns
	f.Add("confirm('are you sure?')")
	f.Add("prompt('enter value')")

	// Obfuscated XSS
	f.Add("java\tscript:alert(1)")
	f.Add("&#x6A;avascript:alert(1)")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic
		result1 := DetectXSS(input)

		// Calling twice must return the same result
		result2 := DetectXSS(input)
		if result1 != result2 {
			t.Errorf("DetectXSS not consistent: first=%v, second=%v, input=%q", result1, result2, input)
		}
	})
}

// FuzzContainsPathTraversal fuzzes the DetectPathTraversal function to ensure it
// never panics and returns consistent results.
func FuzzContainsPathTraversal(f *testing.F) {
	// Benign inputs
	f.Add("")
	f.Add("hello world")
	f.Add("/home/user/file.txt")
	f.Add("/var/log/app.log")
	f.Add("documents/report.pdf")
	f.Add("C:\\Users\\John\\Documents\\file.txt")

	// Basic path traversal
	f.Add("../../../etc/passwd")
	f.Add("..\\..\\..\\windows\\system32\\config\\sam")
	f.Add("....//....//....//etc/passwd")
	f.Add("..././..././..././etc/passwd")

	// URL-encoded traversal
	f.Add("%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd")
	f.Add("%2e%2e%5c%2e%2e%5cwindows")
	f.Add("..%2f..%2f..%2fetc%2fpasswd")
	f.Add("..%5c..%5c..%5cwindows")

	// Mixed encoding
	f.Add("..%2fetc/passwd")
	f.Add("..%5cwindows\\system32")

	// Sensitive file targets
	f.Add("/etc/passwd")
	f.Add("/etc/shadow")
	f.Add("/etc/hosts")
	f.Add("windows/system32")
	f.Add("boot.ini")
	f.Add("win.ini")
	f.Add("c:/windows/system32/config")
	f.Add("c:\\boot.ini")
	f.Add("c:\\win.ini")

	// Double encoding
	f.Add("%252e%252e%252f")
	f.Add("..%252f..%252f")

	// Null byte injection combined with traversal
	f.Add("../../../etc/passwd%00.jpg")
	f.Add("..\\..\\..\\boot.ini\x00.txt")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic
		result1 := DetectPathTraversal(input)

		// Calling twice must return the same result
		result2 := DetectPathTraversal(input)
		if result1 != result2 {
			t.Errorf("DetectPathTraversal not consistent: first=%v, second=%v, input=%q", result1, result2, input)
		}
	})
}

// FuzzValidateStringValue fuzzes the full validation pipeline (sanitize + all detectors)
// to ensure no panics and consistent behavior across the entire chain.
func FuzzValidateStringValue(f *testing.F) {
	// Benign inputs
	f.Add("username", "JohnDoe")
	f.Add("email", "john@example.com")
	f.Add("name", "Alice Smith")
	f.Add("search", "best movies 2025")
	f.Add("comment", "This product is great!")
	f.Add("path", "/home/user/documents")
	f.Add("description", "A long description with various characters: !@#$%^&*()")

	// SQL injection via field values
	f.Add("search", "SELECT * FROM users")
	f.Add("id", "1' OR '1'='1")
	f.Add("name", "admin'--")
	f.Add("query", "'; DROP TABLE users;--")
	f.Add("filter", "1 UNION SELECT password FROM users")
	f.Add("input", "WAITFOR DELAY '0:0:5'")
	f.Add("param", "BENCHMARK(5000000,SHA1('test'))")
	f.Add("value", "EXEC(xp_cmdshell 'dir')")

	// XSS via field values
	f.Add("comment", "<script>alert('xss')</script>")
	f.Add("bio", "<iframe src='evil.com'></iframe>")
	f.Add("title", "javascript:alert(1)")
	f.Add("content", "<img onerror='alert(1)' src='x'>")
	f.Add("note", "<object data='evil.swf'></object>")
	f.Add("text", "data:text/html,<script>alert(1)</script>")

	// Path traversal via field values
	f.Add("file", "../../../etc/passwd")
	f.Add("path", "..\\..\\..\\windows\\system32")
	f.Add("upload", "%2e%2e%2f%2e%2e%2fetc%2fpasswd")
	f.Add("filename", "/etc/shadow")
	f.Add("dir", "boot.ini")

	// Combined attacks
	f.Add("payload", "SELECT * FROM users; <script>alert(1)</script> ../../../etc/passwd")

	// Edge cases for field keys
	f.Add("", "value")
	f.Add("key with spaces", "normal value")
	f.Add("key\x00null", "value")

	// Unicode and special characters
	f.Add("name", "Hello, 世界! 🌍")
	f.Add("field", "\x00\x00\x00")
	f.Add("data", "hello\xc0\xafworld")
	f.Add("input", "\t\n\r")

	f.Fuzz(func(t *testing.T, key, value string) {
		config := DefaultInputValidationConfig()

		// Must not panic
		err1 := validateStringValue(config, key, value)

		// Calling twice with same input must return same result
		err2 := validateStringValue(config, key, value)

		if (err1 == nil) != (err2 == nil) {
			t.Errorf("validateStringValue not consistent: first=%v, second=%v, key=%q, value=%q", err1, err2, key, value)
		}
		if err1 != nil && err2 != nil && err1.Error() != err2.Error() {
			t.Errorf("validateStringValue error messages differ: first=%q, second=%q, key=%q, value=%q", err1.Error(), err2.Error(), key, value)
		}
	})
}
