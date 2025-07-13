package dto

type UploadOrderRequest struct {
	Number string
}

type OrderResponce struct {
	Number     string
	Status     string
	Accrual    float64
	UploadedAt string
}
