package core

// DispatchProbe records one *unauthorized* attempt to POST the receiver's workflow_dispatch
// endpoint and the HTTP status GitHub returned for it.
type DispatchProbe struct {
	Name   string // human label, e.g. "unauthenticated"
	Status int    // HTTP status code GitHub returned
}

// DispatchRejected reports whether an HTTP status means GitHub *refused* the dispatch. The
// authorization-failure codes are 401 (missing/invalid credential), 403 (authenticated but lacks
// actions:write), and 404 (no permission to even see the workflow). Anything else — notably 204,
// the success code for a workflow_dispatch — is NOT a rejection and counts as a security failure.
func DispatchRejected(status int) bool {
	switch status {
	case 401, 403, 404:
		return true
	default:
		return false
	}
}

// LeakedProbes returns the probes that were NOT rejected — i.e. cases where an unauthorized caller
// reached (or triggered) the dispatch endpoint. A non-empty result is a security regression.
func LeakedProbes(probes []DispatchProbe) []DispatchProbe {
	leaked := make([]DispatchProbe, 0)
	for _, p := range probes {
		if !DispatchRejected(p.Status) {
			leaked = append(leaked, p)
		}
	}
	return leaked
}
