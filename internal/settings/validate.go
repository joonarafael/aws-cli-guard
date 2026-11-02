package settings

func validateSettings(path string, doc document) (File, error) {
	out := File{
		Path:    path,
		Version: doc.Version,
		Mode:    0,
		Verify:  false,
	}

	if doc.Identity == nil {
		return out, nil
	}

	if doc.Identity.Verify == nil {
		return File{}, newError(KindInvalid, "", "identity.verify", 0, 0, nil)
	}

	if *doc.Identity.Verify {
		return File{}, newError(KindInvalid, "", "identity.verify", 0, 0, nil)
	}

	return out, nil
}
