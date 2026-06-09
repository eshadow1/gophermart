package utils

import "strings"

// ValidateLuhn проверяет валидность номера по алгоритму Луна
func ValidateLuhn(number string) bool {
	number = strings.TrimSpace(number)
	if number == "" {
		return false
	}

	var sum int
	maxInt := 9
	alternate := false
	for i := len(number) - 1; i >= 0; i-- {
		c := number[i]
		if c < '0' || c > '9' {
			return false
		}
		n := int(c - '0')
		if alternate {
			n *= 2
			if n > maxInt {
				n -= 9
			}
		}
		sum += n
		alternate = !alternate
	}
	return sum%10 == 0
}
