module book_proposals

go 1.23.1

require (
	github.com/go-resty/resty/v2 v2.15.1 // indirect
	marky/openai v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/modelcontextprotocol/go-sdk v1.2.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
)

replace marky/openai => ./openai
