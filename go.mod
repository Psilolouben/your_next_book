module book_proposals

go 1.23.1

require (
	github.com/go-resty/resty/v2 v2.15.1 // indirect
	marky/openai v0.0.0-00010101000000-000000000000
)

require golang.org/x/net v0.29.0 // indirect

replace marky/openai => ./openai
