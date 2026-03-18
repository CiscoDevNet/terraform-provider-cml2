package common

// ErrorString returns err.Error() or an empty string for nil.
func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
