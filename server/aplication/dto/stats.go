package dto

type AdminStatsDTO struct {
	TotalUsers       int64 `json:"total_users"`
	SuspendedUsers   int64 `json:"suspended_users"`
	ActiveOfferings  int64 `json:"active_offerings"`
	OpenLiquidations int64 `json:"open_liquidations"`
	PendingReports   int64 `json:"pending_reports"`
}
