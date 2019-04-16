package mytestpackages

import (
	"fmt"
	"time"
)

//TimeConvert test
func TimeConvert(strTime string) (int64, error) {
	t, err := time.Parse("Mon, 10 Jay 2001 00:01:01 GMT", strTime)
	if err != nil {
		return 0, err
	}

	fmt.Printf("user time: %v, convert time: %v, unix time: %v\n", strTime, t, t.Unix())

	return t.Unix(), nil
}
