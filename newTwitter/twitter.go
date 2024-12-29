package newTwitter

// Basic response structure for follow action
type FollowResponse struct {
	ID         int64  `json:"id"`
	ScreenName string `json:"screen_name"`
	Following  bool   `json:"following"`
	FollowedBy bool   `json:"followed_by"`
}
