package utils

import (
	"flag"
	"fmt"
	"os"
)

var (
	port int
	doc  bool
)

func InitFlag() (int, bool) {
	flag.IntVar(&port, "p", 8080, "指定服务监听的端口")
	flag.BoolVar(&doc, "d", false, "打开 Api 文档")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS]\n", "xyz")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	return port, doc
}
