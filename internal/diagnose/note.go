package diagnose

// Note holds the facts printed for a corrupted file.
type Note struct {
	Class      Class
	Path       string
	Failure    string
	Key        string
	Line       int
	Column     int
	Version    int
	HasVersion bool
	Supported  []int
}
