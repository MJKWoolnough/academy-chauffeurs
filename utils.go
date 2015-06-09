package main

import (
	"errors"
	"time"
)

var ErrInvalidMobileNumber = errors.New("Invalid Mobile Phone Number Format")

type MobileNumber uint64

func (p MobileNumber) String() string {
	var digits [21]byte
	pos := 21
	for num := uint64(p); num > 0; num /= 10 {
		pos--
		digits[pos] = '0' + byte(num%10)
	}
	if pos == 11 {
		pos--
		digits[pos] = '0'
	} else if pos == 10 && digits[10] == '4' && digits[11] == '4' {
		pos--
		digits[pos] = '+'
	} else {
		return "Invalid Mobile Phone Number"
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
		return 0, ErrInvalidMobileNumber
	}
	return num, nil
}

func ValidMobileNumber(pn string) bool {
	_, err := ParseMobileNumber(pn)
	return err == nil
}

func now() int64 {
	return time.Now().Unix()
}
