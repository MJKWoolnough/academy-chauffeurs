package main

import (
	"time"
)

type MobileNumber uint64

func (p MobileNumber) String() string {
	var digits [22]byte
	pos := 22
	for num := uint64(p); num > 0; num /= 10 {
		pos--
		digits[pos] = '0' + byte(num%10)
	}
	/*
		if pos == 11 {
			pos--
			digits[pos] = '0'
		} else if pos == 10 && digits[10] == '4' && digits[11] == '4' {
			pos--
			digits[pos] = '+'
		} else {
			return "Invalid Mobile Phone Number"
		}*/
	if digits[pos] != '0' {
		pos--
		digits[pos] = '+'
	}
	return string(digits[pos:])
}

func ParseMobileNumber(a string) (MobileNumber, error) {
	var num MobileNumber
	for _, digit := range a {
		if digit >= '0' && digit <= '9' {
			num *= 10
			num += MobileNumber(digit - '0')
		}
	}
	if (num < 7000000000 || num >= 8000000000) && (num < 447000000000 || num >= 448000000000) {
		// return 0, ErrInvalidMobileNumber
	}
	return num, nil
}

func now() int64 {
	return time.Now().Unix()
}
