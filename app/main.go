package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"reddio/extra"
	"reddio/models"
	"reddio/pkg/config"
	"reddio/services/delayer"
	"reddio/services/readerFile"
	"reddio/services/reddio"
	"reddio/testInstance"
	"time"
)

var (
	taskId = "c2cf2c1d-cb46-406d-b025-dd6a0036923c"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}

}

func run() error {
	pKeys, err := readerFile.ReadFile("../data/private_keys.txt")
	if err != nil {
		return err
	}

	accs := readerFile.GetAccs(pKeys)

	twitters, err := readerFile.ReadFile("../data/twitter_data.txt")
	if err != nil {
		return err
	}

	twittersData := readerFile.GetTwitters(twitters)

	proxies, err := readerFile.ReadFile("../data/proxies.txt")
	if err != nil {
		return err
	}

	if len(accs) != len(twitters) || len(accs) != len(proxies) || len(twitters) != len(proxies) {
		return fmt.Errorf("%d wallets loaded, %d twitters loaded, %d proxies loaded. Количество кошельков должно быть равно количеству твитеров и количеству прокси.\n",
			len(accs),
			len(twitters),
			len(proxies))
	}

	codes, err := readerFile.ReadFile("../data/register_codes.txt")
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig("../data/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Printf("%d wallets loaded, %d twitters loaded, %d proxies loaded, %d codes loaded\n\n", len(accs), len(twittersData), len(proxies), len(codes))

	for i, acc := range accs {
		codes, err = readerFile.ReadFile("../data/register_codes.txt")
		if err != nil {
			return err
		}

		log.Printf("Account [%d/%d] %s start\n\n", i+1, len(pKeys), acc.Address)

		twitterData := twittersData[i]

		proxyURL, err := url.Parse(fmt.Sprintf("http://%s", proxies[i]))
		if err != nil {
			panic(err)
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		client := http.Client{
			Transport: transport,
		}

		var res bool

		if cfg.Mode == "default" {
			log.Println("Включен дефолтный режим")

			code := codes[rand.Intn(len(codes))]
			log.Printf("Use %s code if user not registered", code)

			res = reddioSequence(client, acc.Address.String(), code, proxies[i], *twitterData, *cfg)
			if err != nil {
				fmt.Println(err)
			}

			if i+1 == len(pKeys) {
				log.Println("Все аккаунты отработаны")
				return nil
			}
		} else if cfg.Mode == "daily" {
			log.Println("Включен режим сбора только дейликов")

			res = reddioSequenceDaily(client, acc.Address.String())
			if err != nil {
				fmt.Println(err)
			}
		}

		if res {
			log.Printf("Account successfully [%d/%d] %s ended\n\n", i+1, len(pKeys), acc.Address)
			delayer.RandomDelay(cfg.DelayBetweenAccs.Min, cfg.DelayBetweenAccs.Max, true)
		} else {
			log.Printf("Account [%d/%d] %s ended with errors\n\n", i+1, len(pKeys), acc.Address)
			delayer.RandomDelay(cfg.DelayBetweenAccs.Min, cfg.DelayBetweenAccs.Max, false)
		}
	}

	return nil
}

func reddioSequence(client http.Client, address, code, proxy string, twitterData models.TwitterData, cfg config.Config) bool {
	userInfo, err := reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Println("Пользователь не зарегистрирован, буду прозводить регистрацию")
			reddio.PreRegister(client, address)
		} else {
			log.Println("Ошибка при получении информации пользователя:", err)
			return false
		}
	}

	if userInfo != nil && userInfo.CheckedIn {
		log.Println("Пользователь уже выполнил ежедневный чекин")
	}

	log.Println("Буду ожидать, перед тем как авторизироваться через твитер")
	delayer.RandomDelay(5, 15, false)

	err = reddio.LoginTwitter(client, address, twitterData, cfg)
	if err != nil {
		log.Println(err)
		return false
	}

	log.Println("Буду ожидать, перед тем как выполнять репост поста")
	delayer.RandomDelay(1, 5, false)
	err = MakeRepost(twitterData.AuthToken, proxy)
	if err != nil {
		log.Println(err)
		return false
	}

	delayer.RandomDelay(cfg.DelayBeforeLogin.Min, cfg.DelayBeforeLogin.Max, false)
	err = reddio.Register(client, address, code)
	if err != nil {
		log.Println(err)
		return false
	}

	delayer.RandomDelay(cfg.DelayBeforeDaily.Min, cfg.DelayBeforeDaily.Max, false)
	err = reddio.DailyCheckIn(client, address)
	if err != nil {
		log.Println(err)
		return false
	}

	delayer.RandomDelay(cfg.DelayBeforeRepost.Min, cfg.DelayBeforeRepost.Max, false)

	err = reddio.VerifyTask(client, address, taskId)
	if err != nil {
		log.Println(err)
		return false
	}

	time.Sleep(time.Second * 10)

	log.Println("Ожидаю 10 секунд после подтверждения задания, чтобы отобразить кол-во поинтов")

	userInfo, err = reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Println("Пользователь не зарегистрирован, буду прозводить регистрацию")
			reddio.PreRegister(client, address)
		} else {
			log.Println("Ошибка при получении информации пользователя:", err)
			return false
		}
	}

	var msg string

	if userInfo.CheckedIn == false {
		msg = "Нет"
	} else {
		msg = "Да"
	}

	log.Printf("Выполнял ли сегодня дейлик: %s, кол-во daily check in %d, кол-во поинтов: %d\n", msg, userInfo.CheckinCount, userInfo.Points)

	return true
}

func reddioSequenceDaily(client http.Client, address string) bool {
	err := reddio.DailyCheckIn(client, address)
	if err != nil {
		log.Println(err)
		return false
	}

	userInfo, err := reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Println("Пользователь не зарегистрирован, буду прозводить регистрацию")
			reddio.PreRegister(client, address)
		} else {
			log.Println("Ошибка при получении информации пользователя:", err)
			return false
		}
	}
	
	status, err := readerFile.AddCodeIfNotExists(userInfo.InviteCode)
	if err != nil {
		log.Printf("Ошибка при добавлении кода: %v", err)
	} else {
		log.Println(status)
	}

	var msg string

	if userInfo.CheckedIn == false {
		msg = "Нет"
	} else {
		msg = "Да"
	}

	log.Printf("Выполнял ли сегодня дейлик: %s, кол-во daily check in %d, кол-во поинтов: %d\n", msg, userInfo.CheckinCount, userInfo.Points)

	return true
}

func MakeRepost(authToken, proxy string) error {
	config := extra.ReadConfig()
	queryIDs := extra.ReadQueryIDs()

	twitter := testInstance.Twitter{}

	ok, reason, logs := twitter.InitTwitter(1, authToken, proxy, config, queryIDs)
	fmt.Println(ok,
		append(logs, "Failed to initialize Twitter client: "+reason))

	success, errorType, followLogs := twitter.Follow("reddio_com")
	fmt.Println(success, errorType, followLogs)

	success, errorType, retweetLogs := twitter.Retweet("https://twitter.com/intent/retweet?tweet_id=1868631594543755535")

	fmt.Println(success, errorType, retweetLogs)

	return nil
}
