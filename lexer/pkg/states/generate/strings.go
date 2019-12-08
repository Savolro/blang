package generate

// Contains returns true if string list contains a string
func Contains(list []string, s string) bool {
	for _, str := range list {
		if str == s {
			return true
		}
	}
	return false
}
