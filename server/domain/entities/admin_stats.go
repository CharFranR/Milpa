package domain

type AdminStats struct {
	TotalUsers       int64
	SuspendedUsers   int64
	ActiveOfferings  int64
	OpenLiquidations int64
	PendingReports   int64
}
