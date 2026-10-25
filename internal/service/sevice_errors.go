package service

type ServiceError struct {
	Code    int
	Message string
}

func (e ServiceError) HttpStatusCode() int {
	return e.Code
}

func (e ServiceError) Error() string {
	return e.Message
}

func NotFoundError(message string) ServiceError {
	return ServiceError{
		Code:    404,
		Message: message,
	}
}

func AlreadyExistsError(message string) ServiceError {
	return ServiceError{
		Code:    409,
		Message: message,
	}
}

func GenericServiceError(message string) ServiceError {
	return ServiceError{
		Code:    500,
		Message: message,
	}
}
