package dto

type Pagination struct {
	Limit  int `json:"limit" schema:"limit,default:100" validate:"gte=1,lte=100"`
	Offset int `json:"offset" schema:"limit,default:0" validate:"lte=10000"`
}
