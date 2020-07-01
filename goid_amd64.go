// +build amd64

package goc

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

var offset = func() int64 {
	var offsetDict = map[string]int64{
		"go1.4":  128,
		"go1.5":  184,
		"go1.6":  192,
		"go1.7":  192,
		"go1.8":  192,
		"go1.9":  152,
		"go1.10": 152,
		"go1.11": 152,
		"go1.12": 152,
		"go1.13": 152,
		"go1.14": 152,
	}
	ver := strings.Join(strings.Split(runtime.Version(), ".")[:2], ".")
	return offsetDict[ver]
}()

func Goid() int64
