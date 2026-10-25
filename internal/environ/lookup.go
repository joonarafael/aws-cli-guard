package environ

import "strings"

// Lookup returns the value for key in env. The second result is true when key appears in env,
// including when its value is empty. Lookup matches os.Getenv first-wins semantics.
func Lookup(key string, env []string) (string, bool) {
	prefix := key + "="
	for _, entry := range env {
		if entry == key {
			return "", true
		}

		if strings.HasPrefix(entry, prefix) {
			return entry[len(prefix):], true
		}
	}

	return "", false
}
