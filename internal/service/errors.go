package billing

func NewError(code string, err string, httpStatusCode int) error {
	return CustomError{code, err, httpStatusCode}
}

type CustomError struct {
	code           string
	err            string
	httpStatusCode int
}

func (e CustomError) Error() string {
	return e.err
}

func (e CustomError) Code() string {
	return e.code
}

func (e CustomError) HTTPStatusCode() int {
	return e.httpStatusCode
}
