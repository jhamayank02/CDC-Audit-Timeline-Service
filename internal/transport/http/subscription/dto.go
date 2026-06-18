package subscriptionhttp

type CreateSubscriptionRequest struct {
	UserID    string `json:"user_id" binding:"required,uuid"`
	PlanName  string `json:"plan_name" binding:"required,oneof=basic plus pro"`
	Status    string `json:"status" binding:"required,oneof=active inactive cancelled"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	AutoRenew *bool  `json:"auto_renew"`
}

type UpdateSubscriptionRequest struct {
	Status    string `json:"status" binding:"omitempty,oneof=active inactive cancelled"`
	AutoRenew *bool  `json:"auto_renew"`
}
