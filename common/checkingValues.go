package common

/*
* Модуль содержащий набор функций для верификации значений
*
* Версия 0.2, дата релиза 19.03.2019
* */

import (
	"regexp"
)

//CheckStringIP проверка ip адреса переданного в виде строки
func CheckStringIP(ip string) (bool, error) {
	pattern := "^((25[0-5]|2[0-4]\\d|[01]?\\d\\d?)[.]){3}(25[0-5]|2[0-4]\\d|[01]?\\d\\d?)$"

	rx, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return rx.MatchString(ip), nil
}

//CheckStringToken проверяет токен полученный от пользователя
func CheckStringToken(str string) (bool, error) {
	pattern := "^\\w+$"

	rx, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return rx.MatchString(str), nil
}

//CheckFolders проверяет имена директорий
func CheckFolders(f []string) (bool, error) {
	if len(f) == 0 {
		return false, nil
	}

	pattern := "^(/|_|\\w)+$"
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	for _, v := range f {
		if !rx.MatchString(v) {
			return false, nil
		}
	}

	return true, nil
}
