// +build !wasm

// Package slf4go ...
package slf4go

import "github.com/fatih/color"

var fatalp = color.New(color.FgRed).PrintFunc()
var fatalf = color.New(color.FgRed).PrintfFunc()

var errorp = color.New(color.FgRed).PrintFunc()
var errorf = color.New(color.FgRed).PrintfFunc()

var warnp = color.New(color.FgYellow).PrintFunc()
var warnf = color.New(color.FgYellow).PrintfFunc()

var infop = color.New(color.FgWhite).PrintFunc()
var infof = color.New(color.FgWhite).PrintfFunc()

var debugp = color.New(color.FgCyan).PrintFunc()
var debugf = color.New(color.FgCyan).PrintfFunc()

var tracep = color.New(color.FgBlue).PrintFunc()
var tracef = color.New(color.FgBlue).PrintfFunc()
