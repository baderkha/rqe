package main

import (
	"fmt"

	"github.com/baderkha/rqe"
)

func main() {
	out, err := rqe.Parse("1 AND IF(ASCII(SUBSTRING((SELECT USER()),1,1))>=100,1, BENCHMARK(2000000,MD5(NOW()))) --", func(col string) bool {
		return true
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rqe.DANGEROUS_DEBUG_COMPILE_SQL(out.SQL, out.Args))

}
