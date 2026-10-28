package decode

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
)

func classifyJSONVersionText(text string) (int, bool) {
	if text == "" {
		return 0, false
	}

	if len(text) > 1 && text[0] == '0' {
		return 0, false
	}

	start := 0

	if text[0] == '-' {
		if len(text) == 1 {
			return 0, false
		}

		start = 1
	}

	for i := start; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return 0, false
		}
	}

	parsed, err := strconv.ParseInt(text, 10, 0)
	if err != nil {
		return 0, false
	}

	if parsed < int64(math.MinInt) || parsed > int64(math.MaxInt) {
		return 0, false
	}

	return int(parsed), true
}

func offsetToLineColumn(data []byte, offset int64) jsonKeyPos {
	line := 1
	column := 1

	if offset > int64(len(data)) {
		offset = int64(len(data))
	}

	for i := range offset {
		if data[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return jsonKeyPos{line: line, column: column}
}

func mapJSONTokenError(data []byte, err error, supported []int) *Error {
	if syn, ok := errors.AsType[*json.SyntaxError](err); ok {
		pos := offsetToLineColumn(data, syn.Offset)

		return newDocError(KindSyntax, ErrSyntax, "", pos.line, pos.column, 0, false, supported)
	}

	if errors.Is(err, io.EOF) {
		return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
	}

	return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
}
