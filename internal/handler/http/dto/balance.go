package dto

type WithdrawRequest struct {
	Order string
	Sum   float64
}
type WithdrawResponce struct {
	Order       string
	Sum         float64
	ProcessedAt string
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
