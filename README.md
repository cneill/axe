# axe

axe takes nginx log files and filters the information down to only the fields you need.

## Installing

`go install github.com/cneill/axe`

## Filtering

`Usage: axe [command]`

__Commands:__
- `ips`: IP addresses of requesting clients
- `paths`: Usernames supplied with requests
- `referers`: Referring URLs
- `requests`: Full requests (method + path + HTTP version)
- `times`: Date/time stamps
- `useragents`: User-Agents supplied by clients
