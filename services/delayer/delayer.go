package delayer

import (
	"fmt"
	"math/rand"
	"time"
)

func RandomDelay(min, max float64, inMinutes bool) {
	delayRange := max - min
	randomDelay := min + rand.Float64()*delayRange

	var delayDuration time.Duration
	var unitStr string

	if inMinutes {
		delayDuration = time.Duration(randomDelay * float64(time.Minute))
		unitStr = "минут"
	} else {
		delayDuration = time.Duration(randomDelay * float64(time.Second))
		unitStr = "секунд"
	}

	fmt.Printf("Ожидание выполнения: %.2f %s\n", randomDelay, unitStr)
	time.Sleep(delayDuration)
}
