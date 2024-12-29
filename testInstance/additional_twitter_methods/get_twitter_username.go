package additional_twitter_methods

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reddio/utils"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tlsClient "github.com/bogdanfinn/tls-client"
)

func GetTwitterUsername(index int, httpClient tlsClient.HttpClient, cookieClient *utils.CookieClient, bearerToken string, csrfToken string) (string, string, string, []string) {
	var logs []string
	errorType := "Unknown"

	for i := 0; i < 1; i++ {
		baseURL := "https://api.x.com/graphql/UhddhjWCl-JMqeiG4vPtvw/Viewer"

		// Создаем параметры запроса
		params := url.Values{}
		params.Add("variables", `{"withCommunitiesMemberships":true}`)
		params.Add("features", `{"rweb_tipjar_consumption_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"creator_subscriptions_tweet_preview_api_enabled":true,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true}`)
		params.Add("fieldToggles", `{"isDelegate":false,"withAuxiliaryUserLabels":false}`)

		// Добавляем параметры к URL
		fullURL := baseURL + "?" + params.Encode()

		// Создаем новый запрос с полным URL
		req, err := http.NewRequest(http.MethodGet, fullURL, nil)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to build activate request: %s", err.Error()))
			continue
		}

		req.Header = http.Header{
			"accept":                {"*/*"},
			"accept-encoding":       {"gzip, deflate, br"},
			"authorization":         {bearerToken},
			"cookie":                {cookieClient.CookiesToHeader()},
			"origin":                {"https://twitter.com"},
			"referer":               {"https://twitter.com/"},
			"sec-ch-ua-mobile":      {"?0"},
			"sec-ch-ua-platform":    {`"Windows"`},
			"sec-fetch-dest":        {"empty"},
			"sec-fetch-mode":        {"cors"},
			"sec-fetch-site":        {"same-site"},
			"x-csrf-token":          {csrfToken},
			"x-twitter-active-user": {"no"},
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to do get username request: %s", err.Error()))
			continue
		}
		defer resp.Body.Close()

		cookieClient.SetCookieFromResponse(resp)

		csrfToken, ok := cookieClient.GetCookieValue("ct0")
		if ok != true {
			logs = append(logs, "Failed to get new csrf token")
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logs = append(logs, fmt.Sprintf("Failed to read response body: %s", err.Error()))
			continue
		}

		bodyString := string(bodyBytes)
		// fmt.Println(bodyString)
		if strings.Contains(bodyString, "screen_name") {
			var responseDataUsername getUsernameJSON
			err = json.Unmarshal(bodyBytes, &responseDataUsername)
			if err != nil {
				logs = append(logs, fmt.Sprintf("Failed to do unmarshal in get username response: %s", err.Error()))
				continue
			}
			username := responseDataUsername.Data.Viewer.UserResults.Result.Legacy.ScreenName
			logs = append(logs, fmt.Sprintf("Successfully got username: %s", username))
			return username, csrfToken, "", logs

		} else if strings.Contains(bodyString, "this account is temporarily locked") {
			logs = append(logs, "Account is temporarily locked")
			return "locked", csrfToken, "Locked", logs

		} else if strings.Contains(bodyString, "Could not authenticate you") {
			logs = append(logs, "Could not authenticate you")
			return "failed_auth", csrfToken, "Unauthenticated", logs
		} else {
			logs = append(logs, fmt.Sprintf("Unknown response: %s", bodyString))
		}
	}

	logs = append(logs, "Unable to get twitter username")
	return "", "", errorType, logs
}

type getUsernameJSON struct {
	Data struct {
		Viewer struct {
			HasCommunityMemberships     bool `json:"has_community_memberships"`
			CreateCommunityActionResult struct {
				Typename string `json:"__typename"`
				Reason   string `json:"reason"`
				Message  string `json:"message"`
			} `json:"create_community_action_result"`
			UserFeatures []struct {
				Feature string `json:"feature"`
				Enabled bool   `json:"enabled"`
			} `json:"user_features"`
			UserResults struct {
				Result struct {
					Typename                   string `json:"__typename"`
					ID                         string `json:"id"`
					RestID                     string `json:"rest_id"`
					AffiliatesHighlightedLabel struct {
					} `json:"affiliates_highlighted_label"`
					HasGraduatedAccess bool   `json:"has_graduated_access"`
					IsBlueVerified     bool   `json:"is_blue_verified"`
					ProfileImageShape  string `json:"profile_image_shape"`
					Legacy             struct {
						Following           bool   `json:"following"`
						CanDm               bool   `json:"can_dm"`
						CanMediaTag         bool   `json:"can_media_tag"`
						CreatedAt           string `json:"created_at"`
						DefaultProfile      bool   `json:"default_profile"`
						DefaultProfileImage bool   `json:"default_profile_image"`
						Description         string `json:"description"`
						Entities            struct {
							Description struct {
								Urls []any `json:"urls"`
							} `json:"description"`
						} `json:"entities"`
						FastFollowersCount      int    `json:"fast_followers_count"`
						FavouritesCount         int    `json:"favourites_count"`
						FollowersCount          int    `json:"followers_count"`
						FriendsCount            int    `json:"friends_count"`
						HasCustomTimelines      bool   `json:"has_custom_timelines"`
						IsTranslator            bool   `json:"is_translator"`
						ListedCount             int    `json:"listed_count"`
						Location                string `json:"location"`
						MediaCount              int    `json:"media_count"`
						Name                    string `json:"name"`
						NeedsPhoneVerification  bool   `json:"needs_phone_verification"`
						NormalFollowersCount    int    `json:"normal_followers_count"`
						PinnedTweetIdsStr       []any  `json:"pinned_tweet_ids_str"`
						PossiblySensitive       bool   `json:"possibly_sensitive"`
						ProfileBannerURL        string `json:"profile_banner_url"`
						ProfileImageURLHTTPS    string `json:"profile_image_url_https"`
						ProfileInterstitialType string `json:"profile_interstitial_type"`
						ScreenName              string `json:"screen_name"`
						StatusesCount           int    `json:"statuses_count"`
						TranslatorType          string `json:"translator_type"`
						Verified                bool   `json:"verified"`
						WantRetweets            bool   `json:"want_retweets"`
						WithheldInCountries     []any  `json:"withheld_in_countries"`
					} `json:"legacy"`
					TipjarSettings struct {
					} `json:"tipjar_settings"`
					LegacyExtendedProfile struct {
						Birthdate struct {
							Day            int    `json:"day"`
							Month          int    `json:"month"`
							Year           int    `json:"year"`
							Visibility     string `json:"visibility"`
							YearVisibility string `json:"year_visibility"`
						} `json:"birthdate"`
					} `json:"legacy_extended_profile"`
					IsProfileTranslatable         bool   `json:"is_profile_translatable"`
					SuperFollowsApplicationStatus string `json:"super_follows_application_status"`
					CreatorSubscriptionsCount     int    `json:"creator_subscriptions_count"`
				} `json:"result"`
			} `json:"user_results"`
			EducationFlags         []any `json:"educationFlags"`
			IsTfeRestrictedSession bool  `json:"is_tfe_restricted_session"`
			IsActiveCreator        bool  `json:"is_active_creator"`
			SuperFollowersCount    int   `json:"super_followers_count"`
		} `json:"viewer"`
		IsSuperFollowSubscriber bool `json:"is_super_follow_subscriber"`
	} `json:"data"`
}
