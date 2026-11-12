package settings

func validateSettings(path string, doc document) (File, error) {
	out := File{
		Path:    path,
		Version: doc.Version,
		Mode:    0,
		Verify:  false,
		Level:   LevelTight,
	}

	if doc.Identity != nil {
		if doc.Identity.Verify == nil {
			return File{}, newError(KindInvalid, "", "identity.verify", 0, 0, nil)
		}

		if *doc.Identity.Verify {
			return File{}, newError(KindInvalid, "", "identity.verify", 0, 0, nil)
		}
	}

	if doc.Logging != nil {
		level, err := validateLogLevel(doc.Logging.Level)
		if err != nil {
			return File{}, err
		}

		out.Level = level
	}

	return out, nil
}

func validateLogLevel(level *string) (Level, error) {
	if level == nil || *level == "" {
		return LevelTight, newError(KindInvalid, "", "logging.level", 0, 0, nil)
	}

	switch Level(*level) {
	case LevelTight, LevelVerbose:
		return Level(*level), nil
	default:
		return LevelTight, newError(KindInvalid, "", "logging.level", 0, 0, nil)
	}
}
