package reddio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reddio/models"
	"reddio/pkg/config"
	"reddio/services/delayer"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type UserInfoData struct {
	CheckedIn    bool    `json:"checked_in"`
	CheckinCount int     `json:"checkin_count"`
	Points       float64 `json:"points"`
	TaskPoints   float64 `json:"task_points"`
	InviteCode   string  `json:"invitation_code"`
}

func setCommonHeaders(req *http.Request) {
	headers := map[string]string{
		"accept":             "application/json, text/plain, */*",
		"accept-language":    "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"origin":             "https://points.reddio.com",
		"priority":           "u=1, i",
		"referer":            "https://points.reddio.com/",
		"sec-ch-ua":          `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func UserInfo(client http.Client, address string) (*UserInfoData, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://points-mainnet.reddio.com/v1/userinfo?wallet_address=%s", address), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make get user info request %s", err)
	}
	setCommonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info %s", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body %s", err)
	}

	var response struct {
		Status string       `json:"status"`
		Error  string       `json:"error"`
		Data   UserInfoData `json:"data"`
	}

	if err := json.Unmarshal(bodyText, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s", err)
	}

	log.Println("Успешно получил статус пользователя")

	if response.Error == "User not registered" {
		return nil, errors.New("user not registered")
	}

	return &response.Data, nil
}
func PreRegister(client http.Client, address string) error {
	data := strings.NewReader(fmt.Sprintf(`{"wallet_address":"%s"}`, address))
	req, err := http.NewRequest("POST", "https://points-mainnet.reddio.com/v1/pre_register", data)
	if err != nil {
		return fmt.Errorf("failed to make post pre register request %s", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "it-IT,it;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://points.reddio.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://points.reddio.com/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post pre register %s", err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	var response struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Data   string `json:"data"`
	}

	if err := json.Unmarshal(bodyText, &response); err != nil {
		return fmt.Errorf("failed to unmarshal %s", err)
	}

	log.Println("Успешно получил статус пользователя на пре регистрации")

	return nil
}

func LoginTwitter(client http.Client, address string, twitterData models.TwitterData, cfg config.Config) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://points-mainnet.reddio.com/v1/login/twitter?wallet_address=%s", address), nil)
	if err != nil {
		return fmt.Errorf("failed to make get twitter login request %s", err)
	}
	setCommonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get twitter login %s", err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	var response struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Data   struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyText, &response); err != nil {
		return fmt.Errorf("failed to unmarshal %s", err)
	}

	parsedURL, err := url.Parse(response.Data.URL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s", err)
	}
	query := parsedURL.Query()

	scope := strings.ReplaceAll(query.Get("scope"), " ", "%20")

	newURL := fmt.Sprintf("https://twitter.com/i/api/2/oauth2/authorize?client_id=%s&code_challenge=%s&code_challenge_method=%s&redirect_uri=%s&response_type=%s&scope=%s&state=%s",
		query.Get("client_id"),
		query.Get("code_challenge"),
		query.Get("code_challenge_method"),
		url.QueryEscape(query.Get("redirect_uri")),
		query.Get("response_type"),
		scope,
		query.Get("state"))

	log.Println("Успешно получил ссылку для твитер авторизации")

	MakeAuthorize(client, newURL, twitterData, cfg)

	return nil
}

func MakeAuthorize(client http.Client, newURL string, twitterData models.TwitterData, cfg config.Config) error {
	fmt.Println(newURL, twitterData)

	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		return fmt.Errorf("failed to make get twitter make authorize request %s", err)
	}
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")

	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("cookie", fmt.Sprintf(`auth_token=%s; ct0=%s`, twitterData.AuthToken, twitterData.Ct0))
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", newURL)
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("x-csrf-token", twitterData.Ct0)
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get twitter make authorize %s", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	var response struct {
		AuthCode string `json:"auth_code"`
		AppName  string `json:"app_name"`
	}

	if err := json.Unmarshal(bodyText, &response); err != nil {
		return fmt.Errorf("failed to unmarshal %s", err)
	}

	authCode := response.AuthCode
	log.Println("Успешно получил код авторизации для твитера")

	return MakeAuthTwitter(client, newURL, authCode, twitterData, cfg)
}

func MakeAuthTwitter(client http.Client, newURL, authCode string, twitterData models.TwitterData, cfg config.Config) error {
	data := strings.NewReader(fmt.Sprintf(`approval=true&code=%s`, authCode))

	req, err := http.NewRequest("POST", "https://twitter.com/i/api/2/oauth2/authorize", data)
	if err != nil {
		return fmt.Errorf("failed to make post twitter make auth twitter request %s", err)
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("cookie", fmt.Sprintf(`auth_token=%s; ct0=%s`, twitterData.AuthToken, twitterData.Ct0))
	req.Header.Set("origin", "https://twitter.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", newURL)
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("x-csrf-token", twitterData.Ct0)
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post twitter login %s", err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	var response struct {
		RedirectUrl string `json:"redirect_uri"`
	}

	if err := json.Unmarshal(bodyText, &response); err != nil {
		return fmt.Errorf("failed to unmarshal %s", err)
	}

	redirectUrl := response.RedirectUrl

	log.Println("Успешно произвел авторизацию в твитер аккаунт")

	delayer.RandomDelay(cfg.DelayBeforeLoginTwitter.Min, cfg.DelayBeforeLoginTwitter.Max, false)

	return RedirectUrl(client, redirectUrl)
}

func RedirectUrl(client http.Client, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to make get redirect url request %s", err)
	}
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-language", "de-CH,de;q=0.9")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=0, i")
	req.Header.Set("referer", "https://twitter.com/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get redirect url %s", err)
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}
	log.Println("Успешно произвел возврат на страницу регистрации")

	return nil
}

func Register(client http.Client, address, invCode string) error {
	var stringData = fmt.Sprintf(`{"wallet_address":"%s","invitation_code":"%s"}`, address, invCode)
	data := strings.NewReader(stringData)
	req, err := http.NewRequest("POST", "https://points-mainnet.reddio.com/v1/register", data)
	if err != nil {
		return fmt.Errorf("failed to post register request %s", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://points.reddio.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://points.reddio.com/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post register user %s", err)
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	log.Println("Успешно произвел регистрацию аккаунта")

	return nil
}

func DailyCheckIn(client http.Client, address string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://points-mainnet.reddio.com/v1/daily_checkin?wallet_address=%s", address), nil)
	if err != nil {
		return fmt.Errorf("failed to get daily check in request %s", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "de-CH,de;q=0.9")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("origin", "https://points.reddio.com")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://points.reddio.com/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get daily check in %s", err)
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	log.Println("Успешно произвел daily check in аккаунта")

	return nil
}

func VerifyTask(client http.Client, address, taskId string) error {
	var stringData = fmt.Sprintf(`{"task_uuid":"%s","wallet_address":"%s"}`, taskId, address)
	data := strings.NewReader(stringData)
	req, err := http.NewRequest("POST", "https://points-mainnet.reddio.com/v1/points/verify", data)
	if err != nil {
		return fmt.Errorf("failed to post verify task request %s", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "de-CH,de;q=0.9")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://points.reddio.com")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://points.reddio.com/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get verify task %s", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body %s", err)
	}

	log.Println("Успешно произвел подтверждения задания за репост")

	fmt.Println(string(body))

	return nil
}

type RequestConfig struct {
	URL                    string
	AcceptHeader           string
	AcceptLanguage         string
	CacheControl           string
	Origin                 string
	Pragma                 string
	Priority               string
	Referer                string
	UserAgent              string
	SecChUa                string
	SecChUaMobile          string
	SecChPlatform          string
	SecFetchDest           string
	SecFetchMode           string
	SecFetchSite           string
	Cookie                 string
	SecChUaArch            string
	SecChUaBitness         string
	SecChUaFullVersion     string
	SecChUaFullVersionList string
	SecChUaModel           string
	SecChUaPlatform        string
	SecChUaPlatformVersion string
}

func (c RequestConfig) getHeaders() fhttp.Header {
	headers := fhttp.Header{}
	headers.Set("accept", c.AcceptHeader)
	headers.Set("accept-language", c.AcceptLanguage)
	headers.Set("cache-control", c.CacheControl)
	headers.Set("origin", c.Origin)
	headers.Set("pragma", c.Pragma)
	headers.Set("priority", c.Priority)
	headers.Set("referer", c.Referer)
	headers.Set("sec-ch-ua", c.SecChUa)
	headers.Set("sec-ch-ua-mobile", c.SecChUaMobile)
	headers.Set("sec-ch-ua-platform", c.SecChPlatform)
	headers.Set("sec-fetch-dest", c.SecFetchDest)
	headers.Set("sec-fetch-mode", c.SecFetchMode)
	headers.Set("sec-fetch-site", c.SecFetchSite)
	headers.Set("user-agent", c.UserAgent)
	headers.Set("sec-ch-ua-arch", c.SecChUaArch)
	headers.Set("sec-ch-ua-bitness", c.SecChUaBitness)
	headers.Set("sec-ch-ua-full-version", c.SecChUaFullVersion)
	headers.Set("sec-ch-ua-full-version-list", c.SecChUaFullVersionList)
	headers.Set("sec-ch-ua-model", c.SecChUaModel)
	headers.Set("sec-ch-ua-platform", c.SecChUaPlatform)
	headers.Set("sec-ch-ua-platform-version", c.SecChUaPlatformVersion)
	return headers
}

func DefaultReddioConfig() RequestConfig {
	return RequestConfig{
		URL:                    "https://testnet-faucet.reddio.com/api/claim/health",
		AcceptHeader:           "application/json, text/plain, */*",
		AcceptLanguage:         "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		CacheControl:           "no-cache",
		Origin:                 "https://testnet-faucet.reddio.com",
		Pragma:                 "no-cache",
		Priority:               "u=1, i",
		Referer:                "https://testnet-faucet.reddio.com/",
		UserAgent:              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		SecChUa:                `"Not A(Brand";v="8", "Chromium";v="132", "Google Chrome";v="132"`,
		SecChUaArch:            `"x86"`,
		SecChUaBitness:         `"64"`,
		SecChUaFullVersion:     `"132.0.6834.160"`,
		SecChUaFullVersionList: `"Not A(Brand";v="8.0.0.0", "Chromium";v="132.0.6834.160", "Google Chrome";v="132.0.6834.160"`,
		SecChUaMobile:          "?0",
		SecChUaModel:           `""`,
		SecChUaPlatform:        `"Windows"`,
		SecChUaPlatformVersion: `"10.0.0"`,
		SecFetchDest:           "empty",
		SecFetchMode:           "cors",
		SecFetchSite:           "same-origin",
	}
}

func GetHealth(proxy string) (string, error) {
	fmt.Println(proxy)
	config := DefaultReddioConfig()

	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_131),
		tls_client.WithCookieJar(jar),
	}

	if proxy != "" {
		options = append(options, tls_client.WithProxyUrl(proxy))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return "", fmt.Errorf("ошибка создания клиента: %v", err)
	}

	req, err := fhttp.NewRequest(http.MethodPost, config.URL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header = config.getHeaders()
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Cookie", config.Cookie)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("неуспешный статус ответа: %d, тело: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
