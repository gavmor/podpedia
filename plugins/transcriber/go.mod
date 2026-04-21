module github.com/gavmor/podpedia/plugins/transcriber

go 1.26.1

require (
	github.com/gavmor/podpedia/gen v0.0.0
	github.com/samber/lo v1.53.0
	go.bytecodealliance.org/cm v0.3.0
)

replace github.com/gavmor/podpedia/gen => ../../gen

require golang.org/x/text v0.22.0 // indirect
