// Command meow provides the CLI for compiling and running .nyan files.
//
// # Usage
//
//	meow run <file.nyan>              Run a .nyan file
//	meow build <file.nyan> [-o name]  Build a binary
//	meow transpile <file.nyan>        Show generated Go code
//	meow test [files...]              Run _test.nyan files
//	meow version                      Show version info
//	meow help [command]               Show help for a command
//	meow <file.nyan>                  Shorthand for 'meow run'
//
// # Flags
//
//	--verbose, -v    Enable debug logging
package main
