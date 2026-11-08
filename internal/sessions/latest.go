package sessions

import "slices"

// LatestForSession returns the last record for sessionID in file order.
func LatestForSession(file File, sessionID string) (Record, bool) {
	for _, v := range slices.Backward(file.Records) {
		if v.SessionID == sessionID {
			return v, true
		}
	}

	var empty Record

	return empty, false
}
