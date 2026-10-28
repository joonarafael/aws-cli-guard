package policy

import (
	"os"
	"time"
)

// Action is a profile or defaults policy action.
type Action string

const (
	// ActionAllow executes without a grant.
	ActionAllow Action = "allow"
	// ActionPrompt requires a valid system-wide grant.
	ActionPrompt Action = "prompt"
	// ActionDeny refuses execution.
	ActionDeny Action = "deny"
)

// Policy is a loaded, validated user policy file.
type Policy struct {
	Path             string
	Version          int
	Mode             os.FileMode
	DefaultAction    Action
	CredentialExport Action
	Profiles         map[string]Profile
}

// Profile holds profile-specific policy settings.
type Profile struct {
	Action         Action
	ApprovalTTL    time.Duration
	HasApprovalTTL bool
}
