package main

import (
	"fmt"

	"github.com/baderkha/rqe"
)

func main() {
	out, err := rqe.Parse(`user_id in age(21)  and user_id in [1,2] and (date_of_birth between age([17,18]) or sat_score gte 1200 and (val eq 4)) or (val eq 1)`, func(col string) bool {
		return true
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rqe.DANGEROUS_DEBUG_COMPILE_SQL(out.SQL, out.Args))

}
