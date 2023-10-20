package errs

// Err implements the Error interface with error marshaling.
type Err struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func New(message, code string) error {
	return &Err{
		Message: message,
		Code:    code,
	}
}

func (e *Err) Error() string {
	return e.Message
}

// IsExpected finds Err{} inside passed error.
func IsExpected(e error) bool {
	_, ok := e.(*Err)
	return ok
}

// GetCode returns code of given error or empty string if error is not custom
func GetCode(err error) string {
	v, ok := err.(*Err)
	if !ok {
		return ""
	}
	return v.Code
}
