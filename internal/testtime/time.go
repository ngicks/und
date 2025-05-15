package testtime

import (
	"time"
	_ "time/tzdata"
)

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

var AsiaTokyo = must(time.LoadLocation("Asia/Tokyo"))

var (
	CurrInUTC       = time.Now().In(time.UTC)
	CurrInAsiaTokyo = CurrInUTC.In(AsiaTokyo)
	OneSecLater     = CurrInUTC.Add(time.Second)
)
