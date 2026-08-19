package models

type DashboardOverview struct {
	TotalLocations     int                      `json:"totalLocations"`
	LocationsThisMonth int                      `json:"locationsThisMonth"`
	ActiveUsers        int                      `json:"activeUsers"`
	UsersThisMonth     int                      `json:"usersThisMonth"`
	PendingReviews     int                      `json:"pendingReviews"`
	ApprovalsToday     int                      `json:"approvalsToday"`
	ApprovalRate       float64                  `json:"approvalRate"`
	PendingSuggestions int                      `json:"pendingSuggestions"`
	WeeklyActivity     []map[string]interface{} `json:"weeklyActivity"`
	RecentActivities   []map[string]interface{} `json:"recentActivities"`
}
