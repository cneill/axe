# axe

axe takes nginx log files and filters the information down to only the fields you need.

## Installing

`go install github.com/cneill/axe`


## Filtering

`Usage: axe [command]`

Pipe logfiles into axe to filter on attributes:

- `ips`: IP addresses of requesting clients
- `paths`: Usernames supplied with requests
- `referers`: Referring URLs
- `requests`: Full requests (method + path + HTTP version)
- `times`: Date/time stamps
- `useragents`: User-Agents supplied by clients
