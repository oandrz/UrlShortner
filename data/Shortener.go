package data

type ShortenerRequest struct {
	Url string `json:"url"`
}

type ShortenerResponse struct {
	Url string `json:"url"`
}
