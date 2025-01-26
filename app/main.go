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
	"reddio/services/exporter"
	"reddio/services/readerFile"
	"reddio/services/reddio"
	"reddio/shuffle"
	"reddio/testInstance"
	"time"
)

var (
	taskId         = "c2cf2c1d-cb46-406d-b025-dd6a0036923c"
	bridgeTaskId   = "c2cf2c1d-cb46-406d-b025-dd6a00369216"
	transferTaskId = "c2cf2c1d-cb46-406d-b025-dd6a00369215"
	faucetTaskId   = "c2cf2c1d-cb46-406d-b025-dd6a00369214"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {
	cfg, err := config.LoadConfig("../data/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if cfg.IsShuffle {
		fmt.Println("Включен режим перемешивания кошельков")

		err = shuffle.ShuffleFiles("../data/private_keys.txt", "../data/twitter_data.txt", "../data/proxies.txt")
		if err != nil {
			return err
		}
	}

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

			res, err = reddioSequence(client, acc.Address.String(), code, proxies[i], *twitterData, *cfg)
			if err != nil {
				fmt.Println(err)
			}

			if i+1 == len(pKeys) {
				log.Println("Все аккаунты отработаны")
				return nil
			}
		} else if cfg.Mode == "daily" {
			log.Println("Включен режим сбора только дейликов")

			//healthCheckProxy := fmt.Sprintf("http://%s", proxies[i])
			//result, err := reddio.GetHealth(healthCheckProxy)
			//if err != nil {
			//	fmt.Println(err)
			//} else {
			//	fmt.Println(result)
			//}
			//
			//ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			//defer cancel()
			//
			//turnstileData, err := captcha.GetTurnstileData(ctx)
			//if err != nil {
			//	fmt.Printf("Error getting Turnstile data: %v\n", err)
			//} else {
			//	fmt.Printf("Turnstile data: %+v\n", turnstileData)
			//}

			res, err = reddioSequenceDaily(client, acc.Address.String(), *cfg)
			if err != nil {
				fmt.Println(err)
			}
		} else if cfg.Mode == "csv" {
			log.Println("Включен режим CSV")

			var allAddresses []string
			for _, account := range accs {
				allAddresses = append(allAddresses, account.Address.String())
			}

			log.Printf("Собрано %d адресов для обработки\n", len(allAddresses))

			res = ProcessAndExportUserData(client, allAddresses, *cfg) == nil
			if err != nil {
				fmt.Println(err)
			}

			if res {
				log.Printf("Успешно обработано и экспортировано %d адресов\n", len(allAddresses))
			} else {
				log.Printf("Произошла ошибка при обработке адресов\n")
			}
			return nil
		}

		if res {
			log.Printf("Account successfully [%d/%d] %s ended\n\n", i+1, len(pKeys), acc.Address)
			switch cfg.Mode {
			case "csv":
				delayer.RandomDelay(cfg.DelayBetweenAccsIfCsv.Min, cfg.DelayBetweenAccsIfCsv.Max, false)
			case "default", "daily":
				delayer.RandomDelay(cfg.DelayBetweenAccs.Min, cfg.DelayBetweenAccs.Max, true)
			}
		} else {
			log.Printf("Account [%d/%d] %s ended with errors\n\n", i+1, len(pKeys), acc.Address)
			switch cfg.Mode {
			case "csv":
				delayer.RandomDelay(cfg.DelayBetweenAccsIfCsv.Min, cfg.DelayBetweenAccsIfCsv.Max, false)
			case "default", "daily":
				delayer.RandomDelay(cfg.DelayBetweenAccs.Min, cfg.DelayBetweenAccs.Max, false)
			}
		}
	}

	return nil
}

func reddioSequence(client http.Client, address, code, proxy string, twitterData models.TwitterData, cfg config.Config) (bool, error) {
	userInfo, err := reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Println("Пользователь не зарегистрирован, буду прозводить регистрацию")
			err = reddio.PreRegister(client, address)
			if err != nil {
				return false, err
			}
		} else {
			log.Println("Ошибка при получении информации пользователя:", err)
			return false, nil
		}
	}

	if userInfo != nil && userInfo.CheckedIn {
		log.Println("Пользователь уже выполнил ежедневный чекин")
	}

	log.Println("Буду ожидать, перед тем как авторизироваться через твитер")
	delayer.RandomDelay(cfg.DelayBeforeLoginTwitter.Min, cfg.DelayBeforeLoginTwitter.Max, false)

	err = reddio.LoginTwitter(client, address, twitterData, cfg)
	if err != nil {
		log.Println(err)
		return false, err
	}

	log.Println("Буду ожидать, перед тем как выполнять репост поста")
	delayer.RandomDelay(1, 5, false)
	err = MakeRepost(twitterData.AuthToken, proxy)
	if err != nil {
		log.Println(err)
		return false, err
	}

	delayer.RandomDelay(cfg.DelayBeforeLogin.Min, cfg.DelayBeforeLogin.Max, false)
	err = reddio.Register(client, address, code)
	if err != nil {
		log.Println(err)
		return false, err
	}

	delayer.RandomDelay(cfg.DelayBeforeDaily.Min, cfg.DelayBeforeDaily.Max, false)
	err = reddio.DailyCheckIn(client, address)
	if err != nil {
		log.Println(err)
		return false, err
	}

	delayer.RandomDelay(cfg.DelayBeforeRepost.Min, cfg.DelayBeforeRepost.Max, false)

	err = reddio.VerifyTask(client, address, taskId)
	if err != nil {
		log.Println(err)
		return false, err
	}

	time.Sleep(time.Second * 10)

	log.Println("Ожидаю 10 секунд после подтверждения задания, чтобы отобразить кол-во поинтов")

	userInfo, err = reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Println("Пользователь не зарегистрирован, буду прозводить регистрацию")
			err = reddio.PreRegister(client, address)
			if err != nil {
				return false, err
			}
		} else {
			log.Println("Ошибка при получении информации пользователя:", err)
			return false, err
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

	log.Printf("Выполнял ли сегодня дейлик: %s, кол-во daily check in %d, кол-во поинтов: %0.2f\n", msg, userInfo.CheckinCount, userInfo.Points)

	return true, nil
}

func reddioSequenceDaily(client http.Client, address string, cfg config.Config) (bool, error) {
	err := reddio.DailyCheckIn(client, address)
	if err != nil {
		log.Println(err)
		return false, err
	}

	delayer.RandomDelay(cfg.DelayBetweenDailyModules.Min, cfg.DelayBetweenDailyModules.Max, false)

	err = reddio.VerifyTask(client, address, faucetTaskId)
	if err != nil {
		log.Println(err)
		return false, err
	} else {
		log.Println("Успешно получил поинты за карн")
	}

	delayer.RandomDelay(cfg.DelayBetweenDailyModules.Min, cfg.DelayBetweenDailyModules.Max, false)

	err = reddio.VerifyTask(client, address, bridgeTaskId)
	if err != nil {
		log.Println(err)
		return false, err
	} else {
		log.Println("Успешно получил поинты за дейли бридж")
	}

	delayer.RandomDelay(cfg.DelayBetweenDailyModules.Min, cfg.DelayBetweenDailyModules.Max, false)

	err = reddio.VerifyTask(client, address, transferTaskId)
	if err != nil {
		log.Println(err)
		return false, err
	} else {
		log.Println("Успешно получил поинты за дейли трансфер")
	}

	userInfo, err := reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Println("Пользователь не зарегистрирован, буду прозводить регистрацию")
			err = reddio.PreRegister(client, address)
			if err != nil {
				return false, err
			}
		} else {
			log.Println("Ошибка при получении информации пользователя:", err)
			return false, nil
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

	log.Printf("Выполнял ли сегодня дейлик: %s, кол-во daily check in %d, кол-во поинтов: %0.2f\n", msg, userInfo.CheckinCount, userInfo.Points)

	return true, nil
}

func ProcessAndExportUserData(client http.Client, addresses []string, cfg config.Config) error {
	var usersData []exporter.UserInfoData

	for _, address := range addresses {
		userInfo, err := getUserInfoWithRetry(client, address)
		if err != nil {
			log.Printf("Ошибка обработки адреса %s: %v", address, err)
			continue
		}

		userData := exporter.UserInfoData{
			Address:      address,
			CheckedIn:    userInfo.CheckedIn,
			CheckinCount: userInfo.CheckinCount,
			Points:       userInfo.Points,
		}
		usersData = append(usersData, userData)

		delayer.RandomDelay(cfg.DelayBetweenAccsIfCsv.Min, cfg.DelayBetweenAccsIfCsv.Max, false)
	}

	if err := exporter.ExportToCSV(usersData, "../data/users_export.csv"); err != nil {
		return fmt.Errorf("ошибка при экспорте в CSV: %w", err)
	}

	return nil
}

func getUserInfoWithRetry(client http.Client, address string) (*reddio.UserInfoData, error) {
	userInfo, err := reddio.UserInfo(client, address)
	if err != nil {
		if err.Error() == "user not registered" {
			log.Printf("Пользователь %s не зарегистрирован, производим регистрацию", address)
			err = reddio.PreRegister(client, address)
			if err != nil {
				return nil, fmt.Errorf("ошибка при регистрации пользователя: %w", err)
			}
			userInfo, err = reddio.UserInfo(client, address)
			if err != nil {
				return nil, fmt.Errorf("ошибка при повторном получении информации: %w", err)
			}
		} else {
			return nil, fmt.Errorf("ошибка при получении информации пользователя: %w", err)
		}
	}
	return userInfo, nil
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
