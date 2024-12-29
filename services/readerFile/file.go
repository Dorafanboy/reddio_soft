package readerFile

import (
	"bufio"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"os"
	"reddio/models"
	"strings"
)

func ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error from open readerFile: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error from scanning: %s", err)
	}

	return lines, nil
}

func GetAccs(lines []string) []*models.Account {
	accs := make([]*models.Account, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		acc := createAccount(lines[i])
		accs = append(accs, acc)
	}

	return accs
}

func createAccount(line string) *models.Account {
	if len(line) == 42 {
		address := common.HexToAddress(line)
		acc := models.NewAccount(address, nil)
		return acc
	} else if len(line) == 66 {
		if line[0:2] == "0x" {
			line = line[2:]
		}

		privateKey, _ := crypto.HexToECDSA(line)
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		acc := models.NewAccount(address, privateKey)

		return acc
	}

	return nil
}

func GetTwitters(lines []string) []*models.TwitterData {
	twitters := make([]*models.TwitterData, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		twitter, err := createTwitterData(lines[i], int8(i+1))
		if err != nil {
			log.Printf("Ошибка во время получения твитеров: %s", err)
		} else {
			twitters = append(twitters, twitter)
		}
	}

	return twitters
}

func createTwitterData(line string, i int8) (*models.TwitterData, error) {
	if line == "" {
		return &models.TwitterData{}, fmt.Errorf("строка пустая строка ошибки %d", i)
	}

	parts := strings.Split(line, ":")

	if len(parts) != 2 {
		return &models.TwitterData{}, fmt.Errorf("неверный формат строки должен быть один разделитель ':' строка ошибки %d", i)
	}

	if parts[0] == "" || parts[1] == "" {
		return &models.TwitterData{}, fmt.Errorf("одна из частей строки пустая строка ошибки %d", i)
	}

	ct0 := parts[0]
	authToken := parts[1]

	twitterData := models.NewTwitterData(ct0, authToken)

	return twitterData, nil
}
