package converter

type Meta struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
	Total  int  `json:"total"`
}
