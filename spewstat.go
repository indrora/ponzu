package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/xattr"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage:", os.Args[0], " [file]")
		os.Exit(1)
	}
	stat, err := os.Stat(os.Args[1])
	if err != nil {
		fmt.Println("failed to stat", os.Args[1], err)
	}
	spew.Dump(stat)
	list_xattrs(stat.Name())
}

func list_xattrs(file string) {
	fh, err := os.Open(file)
	if err != nil {
		return
	}
	defer fh.Close()
	attrs, err := xattr.FList(fh)
	if err != nil {
		return
	}
	for _, attrname := range attrs {
		value, err := xattr.FGet(fh, attrname)
		if err != nil {
			fmt.Println(attrname, " = ? (couldn't list: ", err, ")")
		} else {
			fmt.Println(attrname, "=", value)
		}
	}
}
