package environ_test

func lookupCases() []struct {
	name    string
	key     string
	env     []string
	wantVal string
	wantOK  bool
} {
	return []struct {
		name    string
		key     string
		env     []string
		wantVal string
		wantOK  bool
	}{
		{
			name:    "missing",
			key:     "HOME",
			env:     []string{"PATH=/bin"},
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "empty value present",
			key:     "EMPTY",
			env:     []string{"EMPTY="},
			wantVal: "",
			wantOK:  true,
		},
		{
			name:    "bare key",
			key:     "FLAG",
			env:     []string{"FLAG"},
			wantVal: "",
			wantOK:  true,
		},
		{
			name:    "first wins",
			key:     "X",
			env:     []string{"X=first", "X=second"},
			wantVal: "first",
			wantOK:  true,
		},
		{
			name:    "value with equals",
			key:     "URL",
			env:     []string{"URL=https://example.com?a=b"},
			wantVal: "https://example.com?a=b",
			wantOK:  true,
		},
		{
			name:    "prefix not key",
			key:     "HOME",
			env:     []string{"HOME_DIR=/tmp"},
			wantVal: "",
			wantOK:  false,
		},
	}
}
