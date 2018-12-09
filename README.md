# About
Takes a domain via API and returns registration date. All Whois replies are cached in a SQLite database to minimize outbound request traffic.

# How to Use
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
