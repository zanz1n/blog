package dto

type Pagination struct {
	Limit    int       `json:"limit" schema:"limit,default:100" validate:"gte=1,lte=100"`
	LastSeen Snowflake `json:"last_seen" schema:"last_seen,default:0"`
}
