package shadow

// GatePct exposes the agreement gate for a path so the plain decide report uses the same bar.
func GatePct(path string) (int, bool) { return gatePct(path) }
