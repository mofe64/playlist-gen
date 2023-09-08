package responses

type SpotifyAuthErrorReponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	StatusCode       int    `json:"status_code,omitempty"`
}
type SpotifyOperationErrorResponse struct {
	Error ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
