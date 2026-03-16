package common

func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
