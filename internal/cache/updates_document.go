package cache

import "time"

type updatesDocument struct {
	Version   int     `json:"version"`
	CheckedAt *string `json:"checked_at"`
	Current   *string `json:"current"`
	Latest    *string `json:"latest"`
	Notice    *string `json:"notice"`
}

type updatesEncodeDocument struct {
	Version   int    `json:"version"`
	CheckedAt string `json:"checked_at"`
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Notice    string `json:"notice"`
}

func (r UpdatesRecord) encode() ([]byte, error) {
	doc := updatesEncodeDocument{
		Version:   updatesSchema,
		CheckedAt: r.CheckedAt.UTC().Format(time.RFC3339),
		Current:   r.Current,
		Latest:    r.Latest,
		Notice:    r.Notice,
	}

	return encodeJSONDocument(doc)
}
