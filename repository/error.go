package repository

type BasicAuthError struct {
	Reason string
}

func (e BasicAuthError) Error() (s string) {
	s = "Auth failed. User or password not provided or invalid." + e.Reason
	return
}
