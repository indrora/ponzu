//go:build unix && (aix || ppc64)

package osmeta

// This is a stub for environments that don't have xattrs
func getXattrs(path string) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}
func setXattrs(path string, xat map[string][]byte) error {
	return nil
}
