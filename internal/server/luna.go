package server

import (
	"strconv"
)

func AlgorithmLuna(number string) bool {
	sum := 0
	count := len(number)
	parity := count % 2
	for i := 0; i < count; i++ {
		digit, _ := strconv.Atoi(string(number[i]))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
