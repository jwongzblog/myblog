package main

import (
	"flag"
	"log"
	"src/common"
	"src/runner"
)

var (
	uploadDir   = flag.String("uploadDir", "/mnt/d/code/us3/ufile-gosdk/example", "upload type: dir|file")
	threadCount = flag.Int("threadCount", 10, "thread count")
	product     = flag.String("product", common.Us3, "product: us3 | oss")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	r := runner.NewRunner(*uploadDir, *product, *threadCount)
	r.Run()
}
