package main

import (
	"fmt"

	"github.com/baderkha/rqe"
)

func main() {
	out, err := rqe.Parse(`(someone in []) user_id in ["119","23"] and user_id in [1,2] and (created_at gt '2020-01-01 00:00:00' or sat_score gte 1200 and (val eq 4)) or (val eq 1)`, func(col string) bool {
		return true
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(out.SQL)

}
