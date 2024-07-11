package und

// a script file that vendors option into an internal package.
// Vendoring to avoid cyclic import while the package is being used for internally.

//go:generate go run ./internal/script/vendor_domestic -i ./option -o ./internal/option -e *_test.go,validate_und.go,options.go
