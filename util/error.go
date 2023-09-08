package util

type ApplicationError struct {
	Message string
}

type ApplicationAuthError struct {
	Message string
}

type ApplicationRateLimitError struct {
	Message string
}

func (e ApplicationAuthError) Error() string {
	return e.Message
}

func (e ApplicationRateLimitError) Error() string {
	return e.Message
}

func (e ApplicationError) Error() string {
	return e.Message
}
