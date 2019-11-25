package types

// Go doesn't have a built in min function for integers :(
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
