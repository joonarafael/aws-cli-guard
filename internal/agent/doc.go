// Package agent observes caller context for the aws wrapper.
//
// AG-1 classifies whether standard input is a terminal. AG-2 recognizes callers
// from built-in environment markers when policy enables them. Production catalog
// rows are the pinned product variables in ADR 0022. Recognized callers may
// follow [agents] rules when the policy file defines them (schema version 4).
package agent
