package sessions

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

var (
	sessionIDRE = regexp.MustCompile(`^(sid:[1-9][0-9]*|ppid:[0-9]+)$`)
	effectRE    = regexp.MustCompile(`^(off|on)$`)
)

func validateDocument(path string, doc document) (File, error) {
	if doc.Records == nil {
		return File{}, newError(KindInvalid, "", "records", 0, 0, nil)
	}

	out := File{
		Path:    path,
		Version: doc.Version,
		Mode:    0,
		Records: make([]Record, 0, len(*doc.Records)),
	}

	for i, item := range *doc.Records {
		keyPrefix := "records." + strconv.Itoa(i)

		rec, err := validateRecordItem(item, keyPrefix)
		if err != nil {
			return File{}, stampInvalidError(err, path, doc.Version)
		}

		out.Records = append(out.Records, rec)
	}

	return out, nil
}

func validateRecordItem(item recordDocument, keyPrefix string) (Record, error) {
	if item.SessionID == nil {
		return Record{}, newError(KindInvalid, "", keyPrefix+".session_id", 0, 0, nil)
	}

	if !sessionIDRE.MatchString(*item.SessionID) {
		return Record{}, newError(KindInvalid, "", keyPrefix+".session_id", 0, 0, nil)
	}

	if item.Effect == nil {
		return Record{}, newError(KindInvalid, "", keyPrefix+".effect", 0, 0, nil)
	}

	if !effectRE.MatchString(*item.Effect) {
		return Record{}, newError(KindInvalid, "", keyPrefix+".effect", 0, 0, nil)
	}

	effect := Effect(*item.Effect)

	if item.Token == nil {
		return Record{}, newError(KindInvalid, "", keyPrefix+".token", 0, 0, nil)
	}

	if !validToken(*item.Token) {
		return Record{}, newError(KindInvalid, "", keyPrefix+".token", 0, 0, nil)
	}

	if item.At == nil {
		return Record{}, newError(KindInvalid, "", keyPrefix+".at", 0, 0, nil)
	}

	parsedAt, atErr := time.Parse(time.RFC3339, *item.At)
	if atErr != nil {
		return Record{}, newError(KindInvalid, "", keyPrefix+".at", 0, 0, nil)
	}

	var tty *TTY

	if item.TTY != nil {
		parsed, ttyErr := validateTTY(*item.TTY, keyPrefix+".tty")
		if ttyErr != nil {
			return Record{}, ttyErr
		}

		tty = &parsed
	}

	if effect == EffectOff && tty == nil {
		return Record{}, newError(KindInvalid, "", keyPrefix+".tty", 0, 0, nil)
	}

	return Record{
		SessionID: *item.SessionID,
		Effect:    effect,
		Token:     *item.Token,
		At:        parsedAt.UTC(),
		TTY:       tty,
	}, nil
}

func validateTTY(doc ttyDocument, keyPrefix string) (TTY, error) {
	if doc.Path == nil || *doc.Path == "" {
		return TTY{}, newError(KindInvalid, "", keyPrefix+".path", 0, 0, nil)
	}

	if doc.Dev == nil {
		return TTY{}, newError(KindInvalid, "", keyPrefix+".dev", 0, 0, nil)
	}

	if doc.Ino == nil {
		return TTY{}, newError(KindInvalid, "", keyPrefix+".ino", 0, 0, nil)
	}

	if doc.Rdev == nil {
		return TTY{}, newError(KindInvalid, "", keyPrefix+".rdev", 0, 0, nil)
	}

	if *doc.Dev < 0 || *doc.Ino < 0 || *doc.Rdev < 0 {
		return TTY{}, newError(KindInvalid, "", keyPrefix, 0, 0, nil)
	}

	return TTY{
		Path: *doc.Path,
		Dev:  uint64(*doc.Dev),
		Ino:  uint64(*doc.Ino),
		Rdev: uint64(*doc.Rdev),
	}, nil
}

func stampInvalidError(err error, path string, version int) error {
	sessErr := &Error{}

	ok := errors.As(err, &sessErr)
	if !ok {
		return err
	}

	if sessErr.Path == "" {
		sessErr.Path = stripControls(path)
	}

	sessErr.Version = version
	sessErr.HasVersion = true

	sessErr.Supported = append([]int(nil), supportedSessionVersions()...)

	return sessErr
}

func validateRecordList(path string, list []Record) error {
	for i, rec := range list {
		keyPrefix := "records." + strconv.Itoa(i)

		if !sessionIDRE.MatchString(rec.SessionID) {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".session_id", 0, 0, nil),
				path,
				writeVersion,
			)
		}

		if rec.Effect != EffectOff && rec.Effect != EffectOn {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".effect", 0, 0, nil),
				path,
				writeVersion,
			)
		}

		if !validToken(rec.Token) {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".token", 0, 0, nil),
				path,
				writeVersion,
			)
		}

		if rec.At.IsZero() {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".at", 0, 0, nil),
				path,
				writeVersion,
			)
		}

		if rec.Effect == EffectOff && rec.TTY == nil {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".tty", 0, 0, nil),
				path,
				writeVersion,
			)
		}
	}

	return nil
}
