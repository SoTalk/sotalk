package request

// GetMediaListRequest is the HTTP request for getting media list
type GetMediaListRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}
