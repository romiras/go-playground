module github.com/romiras/googletrans-client

require (
	github.com/mind1949/googletrans v0.0.0
	golang.org/x/text v0.10.0
)

go 1.19

replace github.com/mind1949/googletrans => ../_forks/googletrans
