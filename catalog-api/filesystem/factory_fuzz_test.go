package filesystem

import (
	"testing"
)

func FuzzGetStringSetting(f *testing.F) {
	// Seed corpus entries
	f.Add("host", "default_value")
	f.Add("port", "")
	f.Add("", "fallback")
	f.Add("nonexistent_key", "default")
	f.Add("key with spaces", "val")
	f.Add("\x00", "def")
	f.Add("very_long_key_name_that_might_cause_issues", "x")

	f.Fuzz(func(t *testing.T, key, defaultValue string) {
		// Test 1: empty map always returns default
		result := getStringSetting(nil, key, defaultValue)
		if result != defaultValue {
			t.Errorf("getStringSetting(nil, %q, %q) = %q, want default", key, defaultValue, result)
		}

		// Test 2: empty map
		result = getStringSetting(map[string]interface{}{}, key, defaultValue)
		if result != defaultValue {
			t.Errorf("getStringSetting(empty, %q, %q) = %q, want default", key, defaultValue, result)
		}

		// Test 3: map with the key set to a string value
		m := map[string]interface{}{key: "test_value"}
		result = getStringSetting(m, key, defaultValue)
		if result != "test_value" {
			t.Errorf("getStringSetting(m[%q]=test_value, %q, %q) = %q, want test_value", key, key, defaultValue, result)
		}

		// Test 4: map with the key set to a non-string value returns default
		m2 := map[string]interface{}{key: 12345}
		result = getStringSetting(m2, key, defaultValue)
		if result != defaultValue {
			t.Errorf("getStringSetting(m[%q]=12345, %q, %q) = %q, want default", key, key, defaultValue, result)
		}
	})
}

func FuzzGetIntSetting(f *testing.F) {
	f.Add("port", 445)
	f.Add("", 0)
	f.Add("timeout", -1)
	f.Add("max_value", 2147483647)
	f.Add("\x00null", 21)

	f.Fuzz(func(t *testing.T, key string, defaultValue int) {
		// Test 1: nil map returns default
		result := getIntSetting(nil, key, defaultValue)
		if result != defaultValue {
			t.Errorf("getIntSetting(nil, %q, %d) = %d, want default", key, defaultValue, result)
		}

		// Test 2: empty map returns default
		result = getIntSetting(map[string]interface{}{}, key, defaultValue)
		if result != defaultValue {
			t.Errorf("getIntSetting(empty, %q, %d) = %d, want default", key, defaultValue, result)
		}

		// Test 3: map with int value
		m := map[string]interface{}{key: 9999}
		result = getIntSetting(m, key, defaultValue)
		if result != 9999 {
			t.Errorf("getIntSetting(m[%q]=9999, %q, %d) = %d, want 9999", key, key, defaultValue, result)
		}

		// Test 4: map with float64 value (JSON unmarshalling produces float64)
		m2 := map[string]interface{}{key: float64(8080)}
		result = getIntSetting(m2, key, defaultValue)
		if result != 8080 {
			t.Errorf("getIntSetting(m[%q]=8080.0, %q, %d) = %d, want 8080", key, key, defaultValue, result)
		}

		// Test 5: map with string value returns default
		m3 := map[string]interface{}{key: "not_a_number"}
		result = getIntSetting(m3, key, defaultValue)
		if result != defaultValue {
			t.Errorf("getIntSetting(m[%q]=string, %q, %d) = %d, want default", key, key, defaultValue, result)
		}
	})
}
