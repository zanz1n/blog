package errutils

type OsError interface {
	OsStatus() int
	error
}

type osStatus struct {
	status int
	error
}

func (e *osStatus) OsStatus() int {
	return e.status
}

func NewOs(err error, status int) OsError {
	return &osStatus{status: status, error: err}
}

func Os(err error) OsError {
	if err, ok := err.(OsError); ok {
		return err
	}
	return &osStatus{status: 1, error: err}
}
