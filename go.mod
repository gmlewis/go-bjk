module github.com/gmlewis/go-bjk

go 1.25.0

require (
	github.com/alecthomas/participle/v2 v2.1.0
	github.com/fogleman/fauxgl v0.0.0-20180524200717-d89117924388
	github.com/gmlewis/advent-of-code-2021 v0.37.0
	github.com/gmlewis/go3d v0.0.4
	github.com/gmlewis/irmf-slicer/v3 v3.5.0
	github.com/google/go-cmp v0.6.0
	github.com/hexops/valast v1.4.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/yuin/gopher-lua v1.1.0
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6
)

require (
	github.com/fogleman/simplify v0.0.0-20170216171241-d32f302d5046 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	mvdan.cc/gofumpt v0.4.0 // indirect
)

replace github.com/yuin/gopher-lua v1.1.0 => github.com/gmlewis/gopher-lua v0.0.2
