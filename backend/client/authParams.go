package client

type AuthClientRequestParams struct {
	Method         string
	Path           string
	RequestBody    any
	ExpectedStatus int
	RequestResult  any
	RequestHeaders map[string]string
}
