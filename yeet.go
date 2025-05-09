// Package yeet contains the version number of Yeet.
package yeet

// Version is the current version of yeet.
//
// This variable is set at build time using the -X linker flag. If not set,
// it defaults to "devel".
var Version = "devel"

// BuildMethod contains the method used to build the yeet binary.
//
// This variable is set at build time using the -X linker flag. If not set,
// it defaults to "go-build".
var BuildMethod = "go-build"
