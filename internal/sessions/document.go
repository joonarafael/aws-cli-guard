package sessions

type document struct {
	Version int               `json:"version"`
	Records *[]recordDocument `json:"records"`
}

type recordDocument struct {
	SessionID *string      `json:"session_id"`
	Effect    *string      `json:"effect"`
	Token     *string      `json:"token"`
	At        *string      `json:"at"`
	TTY       *ttyDocument `json:"tty"`
}

type ttyDocument struct {
	Path *string `json:"path"`
	Dev  *int64  `json:"dev"`
	Ino  *int64  `json:"ino"`
	Rdev *int64  `json:"rdev"`
}

type encodeDocument struct {
	Version int         `json:"version"`
	Records []encodeRec `json:"records"`
}

type encodeRec struct {
	SessionID string     `json:"session_id"`
	Effect    string     `json:"effect"`
	Token     string     `json:"token"`
	At        string     `json:"at"`
	TTY       *encodeTTY `json:"tty,omitempty"`
}

type encodeTTY struct {
	Path string `json:"path"`
	Dev  uint64 `json:"dev"`
	Ino  uint64 `json:"ino"`
	Rdev uint64 `json:"rdev"`
}
