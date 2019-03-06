package common

/*
* Модуль содержащий набор функций для верификации значений
*
* Версия 0.1, дата релиза 06.03.2019
* */

import (
	"fmt"
	"regexp"
)

//CheckStringIP проверка ip адреса переданного в виде строки
func CheckStringIP(ip string) (bool, error) {
	fmt.Println("validate ip address", ip)

	pattern := "^((25[0-5]|2[0-4]\\d|[01]?\\d\\d?)[.]){3}(25[0-5]|2[0-4]\\d|[01]?\\d\\d?)$"

	rx, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return rx.MatchString(ip), nil
}
