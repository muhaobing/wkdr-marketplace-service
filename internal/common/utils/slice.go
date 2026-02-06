package utils

func InStringSlice(target string, slice []string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}
