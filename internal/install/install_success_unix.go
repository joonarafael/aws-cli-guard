//go:build unix

package install

import (
	"aws-cli-guard/internal/output"
	"fmt"
	"io"
)

func writeSuccess(stdout io.Writer, prefix string) {
	output.Detail(stdout, "prefix: "+prefix)
	_, _ = fmt.Fprintf(stdout, "export PATH='%s:$PATH'\n", prefix)
	output.Detail(stdout, "shell: aws-guard shell bash or aws-guard shell zsh")
}
