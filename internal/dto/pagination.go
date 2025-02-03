package dto

type Pagination struct {
	Limit  uint `json:"limit" schema:"limit,default:100" validate:"gte=1,lte=100"`
	Offset uint `json:"offset" schema:"limit,default:0" validate:"lte=10000"`
}
