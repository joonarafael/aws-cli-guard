// Package agent observes caller context for the aws wrapper.
//
// Classify whether standard input is a terminal. Recognize callers from built-in
// environment markers when policy enables them. Production catalog rows are in
// markers_catalog.go. Recognized callers may follow [agents] rules in policy v4.
package agent
