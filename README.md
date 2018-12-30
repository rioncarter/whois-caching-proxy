# About
Takes a domain via API and returns registration date. All Whois replies are cached in a SQLite database to minimize outbound request traffic.

# How to build
Golang 1.11 (or higher) is *required* to build this tool

Steps to compile:

- git clone https://github.com/rioncarter/whois-caching-proxy.git
- cd whois-caching-proxy
- go build

This should produce an executable named `whois-caching-proxy`


# How to run
- Specify a port to bind to (defaults to `:9091`) using -BindPort=9091
- Use the `-VerboseLog=true` flag to see requests/responses to the API endpoint
- Callers will receive a JSON object that looks like this in response:
```
{
  "Name": "domain.tld",
  "RegisteredRaw": "2001-10",
  "Registered": "2001-10",
  "RegisteredDate": "2001-10-01T00:00:00Z"
}
```

For now only the year and month are returned.
