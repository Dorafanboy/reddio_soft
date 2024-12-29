package testInstance

import (
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsClient "github.com/bogdanfinn/tls-client"
	"io"
	"reddio/extra"
	"reddio/newTwitter"
	additional_twitter_methods2 "reddio/testInstance/additional_twitter_methods"
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

func (twitter *Twitter) InitTwitter(index int, authToken string, proxy string, config extra.Config, queryID extra.QueryIDs) (bool, string, []string) {
	var logs []string

	twitter.index = index
	twitter.authToken = authToken
	twitter.proxy = proxy
	twitter.config = config
	twitter.queryID = queryID

	ok, reason, initLogs := twitter.prepareClient()
	logs = append(logs, initLogs...)
	return ok, reason, logs
}

func (twitter *Twitter) prepareClient() (bool, string, []string) {
	var logs []string
	var err error

	for i := 0; i < twitter.config.Info.MaxTasksRetries; i++ {
		twitter.client, err = utils.CreateHttpClient(twitter.proxy)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to create HTTP client: %s", err.Error()))
			continue
		}

		twitter.cookies = utils.NewCookieClient()
		twitter.authToken, twitter.ct0, err = additional_twitter_methods2.SetAuthCookies(twitter.index, twitter.cookies, twitter.authToken)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to set auth cookies: %s", err.Error()))
			continue
		}

		var username, ct0 string
		username, ct0, _, usernameLogs := additional_twitter_methods2.GetTwitterUsername(
			twitter.index,
			twitter.client,
			twitter.cookies,
			twitter.queryID.BearerToken,
			twitter.ct0,
		)
		logs = append(logs, usernameLogs...)

		twitter.Username = username
		twitter.ct0 = ct0

		if twitter.Username == "locked" {
			return false, "locked", logs
		} else if twitter.Username == "failed_auth" {
			return false, "failed_auth", logs
		} else if twitter.Username != "" {
			logs = append(logs, "Client prepared successfully")
			return true, "ok", logs
		}
	}

	logs = append(logs, "Failed to prepare client")
	return false, "unknown", logs
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

func GetTweetID(tweetLink string) string {
	tweetLink = strings.TrimSpace(tweetLink)

	var tweetID string
	if strings.Contains(tweetLink, "tweet_id=") {
		parts := strings.Split(tweetLink, "tweet_id=")
		tweetID = strings.Split(parts[1], "&")[0]
	} else if strings.Contains(tweetLink, "?") {
		parts := strings.Split(tweetLink, "status/")
		tweetID = strings.Split(parts[1], "?")[0]
	} else if strings.Contains(tweetLink, "status/") {
		parts := strings.Split(tweetLink, "status/")
		tweetID = parts[1]
	} else {
		extra.Logger{}.Error("Failed to get tweet ID from your link: %s", tweetLink)
		return ""
	}

	return tweetID
}

type retweetResponse struct {
	Data struct {
		CreateRetweet struct {
			RetweetResults struct {
				Result struct {
					RestID string `json:"rest_id"`
					Legacy struct {
						FullText string `json:"full_text"`
					} `json:"legacy"`
				} `json:"result"`
			} `json:"retweet_results"`
		} `json:"create_retweet"`
	} `json:"data"`
}

type alreadyLikedResponse struct {
	Errors []struct {
		Message   string `json:"message"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"locations"`
		Path       []string `json:"path"`
		Extensions struct {
			Name    string `json:"name"`
			Source  string `json:"source"`
			Code    int    `json:"code"`
			Kind    string `json:"kind"`
			Tracing struct {
				TraceID string `json:"trace_id"`
			} `json:"tracing"`
		} `json:"extensions"`
		Code    int    `json:"code"`
		Kind    string `json:"kind"`
		Name    string `json:"name"`
		Source  string `json:"source"`
		Tracing struct {
			TraceID string `json:"trace_id"`
		} `json:"tracing"`
	} `json:"errors"`
	Data struct {
	} `json:"data"`
}

func (twitter *Twitter) Retweet(tweetLink string) (bool, string, []string) {
	var logs []string
	errorType := "Unknown"

	retweetURL := fmt.Sprintf("https://twitter.com/i/api/graphql/%s/CreateRetweet", twitter.queryID.RetweetID)
	tweetID := GetTweetID(tweetLink)
	if tweetID == "" {
		logs = append(logs, "Invalid tweet link")
		return false, errorType, logs
	}
	fmt.Println(tweetID)

	for i := 0; i < twitter.config.Info.MaxTasksRetries; i++ {
		var stringData = fmt.Sprintf(`{"variables":{"tweet_id":"%s","dark_request":false},"queryId":"%s"}`, tweetID, twitter.queryID.RetweetID)
		data := strings.NewReader(stringData)

		// Create new request
		req, err := http.NewRequest("POST", retweetURL, data)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to build retweet request: %s", err.Error()))
			continue
		}
		req.Header = http.Header{
			"accept":                {"*/*"},
			"accept-encoding":       {"gzip, deflate, br"},
			"authorization":         {twitter.queryID.BearerToken},
			"content-type":          {"application/json"},
			"cookie":                {twitter.cookies.CookiesToHeader()},
			"origin":                {"https://twitter.com"},
			"referer":               {tweetLink},
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
			logs = append(logs, fmt.Sprintf("Failed to do retweet request: %s", err.Error()))
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to read retweet response body: %s", err.Error()))
			continue
		}

		bodyString := string(bodyBytes)

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			if strings.Contains(bodyString, "already") {
				var responseAlreadyLike alreadyLikedResponse
				err = json.Unmarshal(bodyBytes, &responseAlreadyLike)
				if err != nil {
					logs = append(logs, fmt.Sprintf("Failed to unmarshal already retweeted response: %s", err.Error()))
					continue
				}
				logs = append(logs, fmt.Sprintf("%s already retweeted tweet %s", twitter.Username, tweetID))
				return true, "", logs
			} else if strings.Contains(bodyString, "create_retweet") {
				var responseRetweet retweetResponse
				err = json.Unmarshal(bodyBytes, &responseRetweet)
				if err != nil {
					logs = append(logs, fmt.Sprintf("Failed to unmarshal retweeted response: %s", err.Error()))
					continue
				}
				logs = append(logs, fmt.Sprintf("%s retweeted tweet %s", twitter.Username, tweetID))
				return true, "", logs
			}

		} else if strings.Contains(bodyString, "this account is temporarily locked") {
			logs = append(logs, "Account is temporarily locked")
			return false, "Locked", logs

		} else if strings.Contains(bodyString, "Could not authenticate you") {
			logs = append(logs, "Could not authenticate you")
			return false, "Unauthenticated", logs
		} else {
			logs = append(logs, fmt.Sprintf("Unknown response while retweet: %s", bodyString))
			continue
		}
	}

	logs = append(logs, "Unable to do retweet")
	return false, errorType, logs
}
