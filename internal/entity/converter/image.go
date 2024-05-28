package converter

type ImageResponse struct {
	Message string    `json:"message"`
	Data    *ImageUrl `json:"data"`
}

type ImageUrl struct {
	ImageUrl string `json:"imageUrl"`
}
