package utils

import "regexp"

func CheckIsNumbersOnly(s string) bool {
	re := regexp.MustCompile("^[0-9]{2,20}$")
	return re.MatchString(s)
}

func CheckLuhn(s string) bool {
	summ := 0
	skip := true
	for i := len(s) - 1; i >= 0; i-- {
		digit := int(s[i] - '0')
		if !skip {
			skip = true
			x2digit := 2 * digit
			if x2digit > 9 {
				x2digit -= 9
			}
			summ += x2digit
			continue
		}
		skip = false
		summ += digit
	}
	return summ%10 == 0
}
