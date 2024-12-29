package models

import (
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsClient "github.com/bogdanfinn/tls-client"
	"io"
	"reddio/extra"
	"reddio/newTwitter"
	"reddio/utils"
	"strings"
)

type Twitter struct {
	index     int
	authToken string
	proxy     string
	config    extra.Config
	queryID   extra.QueryIDs

	ct0      string
	Username string

	client  tlsClient.HttpClient
	cookies *utils.CookieClient
	logger  extra.Logger
}

// Follow makes a subscription to the given username
func (twitter *Twitter) Follow(usernameToFollow string) (bool, string, []string) {
	var logs []string
	errorType := "Unknown"

	for i := 0; i < twitter.config.Info.MaxTasksRetries; i++ {
		var stringData = fmt.Sprintf(`include_profile_interstitial_type=1&include_blocking=1&include_blocked_by=1&include_followed_by=1&include_want_retweets=1&include_mute_edge=1&include_can_dm=1&include_can_media_tag=1&include_ext_has_nft_avatar=1&include_ext_is_blue_verified=1&include_ext_verified_type=1&include_ext_profile_image_shape=1&skip_status=1&screen_name=%s`, usernameToFollow)
		data := strings.NewReader(stringData)

		// Create new request
		req, err := http.NewRequest("POST", "https://twitter.com/i/api/1.1/friendships/create.json", data)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to build follow request: %s", err.Error()))
			continue
		}

		req.Header = http.Header{
			"accept":                {"*/*"},
			"accept-encoding":       {"gzip, deflate, br"},
			"authorization":         {twitter.queryID.BearerToken},
			"content-type":          {"application/x-www-form-urlencoded"},
			"cookie":                {twitter.cookies.CookiesToHeader()},
			"origin":                {"https://twitter.com"},
			"referer":               {fmt.Sprintf("https://twitter.com/%s", usernameToFollow)},
			"sec-ch-ua-mobile":      {"?0"},
			"sec-ch-ua-platform":    {`"Windows"`},
			"sec-fetch-dest":        {"empty"},
			"sec-fetch-mode":        {"cors"},
			"sec-fetch-site":        {"same-origin"},
			"x-csrf-token":          {twitter.ct0},
			"x-twitter-active-user": {"yes"},
			"x-twitter-auth-type":   {"OAuth2Session"},
			http.HeaderOrderKey: {
				"accept",
				"accept-encoding",
				"authorization",
				"content-type",
				"cookie",
				"origin",
				"referer",
				"sec-ch-ua-mobile",
				"sec-ch-ua-platform",
				"sec-fetch-dest",
				"sec-fetch-mode",
				"sec-fetch-site",
				"user-agent",
				"x-csrf-token",
				"x-twitter-active-user",
				"x-twitter-auth-type",
			},
			http.PHeaderOrderKey: {":authority", ":method", ":path", ":scheme"},
		}

		resp, err := twitter.client.Do(req)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to do follow request: %s", err.Error()))
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to read follow response body: %s", err.Error()))
			continue
		}

		bodyString := string(bodyBytes)

		if strings.Contains(bodyString, "screen_name") && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			var responseDataUsername newTwitter.FollowResponse
			err = json.Unmarshal(bodyBytes, &responseDataUsername)
			if err != nil {
				logs = append(logs, fmt.Sprintf("Failed to unmarshal follow response: %s", err.Error()))
				continue
			}
			logs = append(logs, fmt.Sprintf("%s subscribed to %s", twitter.Username, usernameToFollow))
			return true, "", logs

		} else if strings.Contains(bodyString, "this account is temporarily locked") {
			logs = append(logs, "Account is temporarily locked")
			return false, "Locked", logs

		} else if strings.Contains(bodyString, "Could not authenticate you") {
			logs = append(logs, "Could not authenticate you")
			return false, "Unauthenticated", logs
		} else {
			logs = append(logs, fmt.Sprintf("Unknown response while follow: %s", bodyString))
			continue
		}
	}

	logs = append(logs, "Unable to do follow")
	return false, errorType, logs
}
