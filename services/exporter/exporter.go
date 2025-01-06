package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
)

type UserInfoData struct {
	Address      string `json:"address"`
	CheckedIn    bool   `json:"checked_in"`
	CheckinCount int    `json:"checkin_count"`
	Points       int    `json:"points"`
}

func ExportToCSV(users []UserInfoData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	headers := []string{
		"Wallet Address",
		"Total Points",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("ошибка при записи заголовков: %w", err)
	}

	for _, user := range users {
		row := []string{
			user.Address,
			fmt.Sprintf("%d", user.Points),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("ошибка при записи строки: %w", err)
		}
	}

	return nil
}
