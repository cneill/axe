# axe

axe takes nginx log files and filters the information down to only the fields you need.

## Installing

__go >=1.11:__

`go get github.com/cneill/axe`

__go <1.11:__

`go install github.com/cneill/axe`

## Usage

```
axe takes logfiles as STDIN and prints the information requested
Usage: axe [command] [options]

Commands and options:
help
ips
  -resolve
        resolve IPs to hostnames where possible
paths
requests
referers
statuses
times
  -format string
        specify the format for time output (default "02/Jan/2006:15:04:05 -0700")
user-agents
  -simplify
        simplify user-agents
```

### Examples

__Parse file:__

```bash
axe ips < access.log
```

__Concatenate all log files, gzipped or not, and parse:__

```bash
zcat -f access* | axe ips
```
