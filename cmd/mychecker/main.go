// package mychecker runs some standard and custom analyzers.
package main

import (
	"github.com/TheLuckymadman/metawatch/mychecker"
)

func main() {
	mychecker.RunStandadCheks()
	mychecker.RunStaticCheks()
	mychecker.RunNoExitCheck()
}
