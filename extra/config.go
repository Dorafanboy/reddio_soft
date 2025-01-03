package extra

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	Info struct {
		License                 string `yaml:"license"`
		CapsolverApiKey         string `yaml:"capsolver_api_key"`
		MaxTasksRetries         int    `yaml:"max_tasks_retries"`
		MaxUnfreezeAttempts     int    `yaml:"max_unfreeze_attempts"`
		AutoUnfreeze            bool   `yaml:"auto_unfreeze"`
		PauseBetweenTasksRaw    string `yaml:"pause_between_tasks"`
		PauseBetweenAccountsRaw string `yaml:"pause_between_accounts"`
		PauseBetweenTasks       struct {
			Start int
			End   int
		} `yaml:"-"`
		PauseBetweenAccounts struct {
			Start int
			End   int
		} `yaml:"-"`
		AccountRange struct {
			Start int
			End   int
		} `yaml:"-"`
		AccountRangeRaw string `yaml:"account_range"`
	} `yaml:"info"`
	Proxy struct {
		MobileProxy   bool `yaml:"mobile_proxy"`
		ChangeIPPause int  `yaml:"change_ip_pause"`
	} `yaml:"proxy"`
	Data struct {
		Random       bool   `yaml:"random"`
		TweetChatGPT bool   `yaml:"tweet_chatGPT_generated"`
		ChatGPTToken string `yaml:"chatGPT_token"`
	} `yaml:"data"`
}

func ReadConfig() Config {
	_, filename, _, _ := runtime.Caller(0)
	dirname := filepath.Dir(filename)

	filePath := filepath.Join(dirname, "config.yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	accountRangeParts := strings.Split(config.Info.AccountRangeRaw, "-")
	if len(accountRangeParts) == 2 {
		var start, end int
		_, errStart := fmt.Sscanf(accountRangeParts[0], "%d", &start)
		_, errEnd := fmt.Sscanf(accountRangeParts[1], "%d", &end)
		if errStart == nil && errEnd == nil {
			config.Info.AccountRange.Start = start
			config.Info.AccountRange.End = end
		} else {
			Logger{}.Error("Failed to read account_range from config:", errStart, errEnd)
		}
	} else {
		Logger{}.Error("Wrong account_range format")
	}
	pauseBetweenTasksParts := strings.Split(config.Info.PauseBetweenTasksRaw, "-")
	if len(pauseBetweenTasksParts) == 2 {
		var start, end int
		_, errStart := fmt.Sscanf(pauseBetweenTasksParts[0], "%d", &start)
		_, errEnd := fmt.Sscanf(pauseBetweenTasksParts[1], "%d", &end)
		if errStart == nil && errEnd == nil {
			config.Info.PauseBetweenTasks.Start = start
			config.Info.PauseBetweenTasks.End = end
		} else {
			Logger{}.Error("Failed to read pause_between_tasks from config:", errStart, errEnd)
		}
	} else {
		Logger{}.Error("Wrong pause_between_tasks format")
	}

	pauseBetweenAccountsParts := strings.Split(config.Info.PauseBetweenAccountsRaw, "-")
	if len(pauseBetweenAccountsParts) == 2 {
		var start, end int
		_, errStart := fmt.Sscanf(pauseBetweenAccountsParts[0], "%d", &start)
		_, errEnd := fmt.Sscanf(pauseBetweenAccountsParts[1], "%d", &end)
		if errStart == nil && errEnd == nil {
			config.Info.PauseBetweenAccounts.Start = start
			config.Info.PauseBetweenAccounts.End = end
		} else {
			Logger{}.Error("Failed to read pause_between_accounts from config:", errStart, errEnd)
		}
	} else {
		Logger{}.Error("Wrong pause_between_accounts format")
	}

	return config
}

func ReadQueryIDs() QueryIDs {
	config := QueryIDs{
		BearerToken:      "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		LikeID:           "lI07N6Otwv1PhnEgXILM7A",
		RetweetID:        "ojPdsZsimiJrUGLR1sjUtA",
		TweetID:          "bDE2rBtZb3uyrczSZ_pI9g",
		UnlikeID:         "ZYKSe-w7KEslx3JhSIk5LA",
		UnretweetID:      "iQtK4dl5hBmXewYZuEOKVw",
		ScrapeRetweeters: "0BoJlKAxoNPQUHRftlwZ2w",
		ScrapeLikes:      "XRRjv1-uj1HZn3o324etOQ",
	}
	return config
}

type QueryIDs struct {
	BearerToken      string `yaml:"bearer_token"`
	LikeID           string `yaml:"like"`
	RetweetID        string `yaml:"retweet"`
	TweetID          string `yaml:"tweet"`
	UnlikeID         string `yaml:"unlike"`
	UnretweetID      string `yaml:"unretweet"`
	ScrapeRetweeters string `yaml:"scrape_retweeters"`
	ScrapeLikes      string `yaml:"scrape_likes"`
}
