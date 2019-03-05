package main

const (
	itemError          itemType = iota // 0
	itemSpace                          // 1
	itemInt                            // 2
	itemWord                           // 3
	itemLeftDelimiter                  // 4
	itemRightDelimiter                 // 5
	itemQuotedString                   // 6
	itemIP                             // 7

	// itemEOF                         // 1 - UNUSED
	// itemNewline                     // 3 - UNUSED
	// itemNumber                      // 4 - UNUSED

)

const (
	digits = "0987654321"
	eof    = rune(0)

	// ValueNil represents a discarded value (if an error is returned while parsing)
	ValueNil = "NIL"
	// ValueBodyBytes represents the number of bytes transferred
	ValueBodyBytes = "BODY_BYTES"
	// ValueIgnore represents an explicitly ignored value
	ValueIgnore = "IGNORE"
	// ValueIP represents the client IP address
	ValueIP = "IP"
	// ValueReferer represents the referring URL, if any
	ValueReferer = "REFERER"
	// ValueRequest represents the request - method, path, HTTP version
	ValueRequest = "REQUEST"
	// ValueStatus represents the HTTP status code returned
	ValueStatus = "STATUS"
	// ValueTime represents the time of the request
	ValueTime = "TIME"
	// ValueUser represents the username supplied, if any
	ValueUser = "USER"
	// ValueUserAgent represents the user-agent supplied, if any
	ValueUserAgent = "USER_AGENT"
)

var nginxTimeFormat = "02/Jan/2006:15:04:05 -0700"

// $ip - $user [$time $tz] "$req" $status $bytes "$ref" "$ua"
var nginxItemOrder = []*ItemParser{
	ParserIP,            // 0
	ParserIgnore,        // 1
	ParserUser,          // 2
	ParserDelimitedTime, // 3
	ParserRequest,       // 4
	ParserStatus,        // 5
	ParserBodyBytes,     // 6
	ParserReferer,       // 7
	ParserUserAgent,     // 8
}
