package shuffle

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func ShuffleFiles(filePath1, filePath2, filePath3 string) error {
	lines1, err := readLines(filePath1)
	if err != nil {
		return fmt.Errorf("error reading first file: %w", err)
	}

	lines2, err := readLines(filePath2)
	if err != nil {
		return fmt.Errorf("error reading second file: %w", err)
	}

	lines3, err := readLines(filePath3)
	if err != nil {
		return fmt.Errorf("error reading third file: %w", err)
	}

	if len(lines1) != len(lines2) || len(lines1) != len(lines3) {
		return fmt.Errorf("files have different number of lines: %d vs %d vs %d",
			len(lines1), len(lines2), len(lines3))
	}

	indices := make([]int, len(lines1))
	for i := range indices {
		indices[i] = i
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	shuffled1 := make([]string, len(lines1))
	shuffled2 := make([]string, len(lines2))
	shuffled3 := make([]string, len(lines3))
	for i, idx := range indices {
		shuffled1[i] = lines1[idx]
		shuffled2[i] = lines2[idx]
		shuffled3[i] = lines3[idx]
	}

	if err := writeLines(filePath1, shuffled1); err != nil {
		return fmt.Errorf("error writing first file: %w", err)
	}
	if err := writeLines(filePath2, shuffled2); err != nil {
		return fmt.Errorf("error writing second file: %w", err)
	}
	if err := writeLines(filePath3, shuffled3); err != nil {
		return fmt.Errorf("error writing third file: %w", err)
	}

	fmt.Println("Кошельки успешно перемешаны")
	fmt.Println()
	return nil
}

func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
