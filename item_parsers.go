package main

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type value struct {
	items     []item
	obj       interface{}
	valueType string
}

func (v value) pos() int {
	if len(v.items) == 0 {
		return 0
	}
	pos := math.MaxInt32
	for _, i := range v.items {
		if i.pos < pos {
			pos = i.pos
		}
	}
	return pos
}

func nilVal(items []item) value {
	return value{
		items:     items,
		obj:       nil,
		valueType: ValueNil,
	}
}

type ipFn func(...item) (value, error)

var errItemCount = fmt.Errorf("invalid number of items")

// ItemParser handles converting items into their field values for addition to *LogLine
type ItemParser struct {
	valueType string
	producers []itemProducer
	parseFn   ipFn
}

func (i *ItemParser) parse(input ...item) (value, error) {
	if len(input) != len(i.producers) {
		return nilVal(input), errItemCount
	}

	if i.parseFn == nil {
		return nilVal(input), nil
	}

	for _, in := range input {
		if in.typ == itemError {
			return nilVal(input), fmt.Errorf("invalid value:%d: %s", in.pos, in.val)
		}
	}

	v, err := i.parseFn(input...)
	if err != nil {
		err = fmt.Errorf("%d:%s: %v", v.pos(), i.valueType, err)
	}
	return v, err
}

// ParserDelimitedTime takes left delimiter, time, timezone, and right delimiter items, producing a time.Time
var ParserDelimitedTime = &ItemParser{
	valueType: ValueTime,
	producers: []itemProducer{leftDelimProducer, wordProducer, wordProducer, rightDelimProducer},
	parseFn:   parseTime,
}

func parseTime(input ...item) (value, error) {
	fullTime := input[1].val + " " + input[2].val
	parsedTime, err := time.Parse(nginxTimeFormat, fullTime)
	if err != nil {
		return nilVal(input), err
	}
	return value{input, parsedTime, ValueTime}, nil
}

// ParserIgnore takes a word item and suppresses its addition
var ParserIgnore = &ItemParser{
	valueType: ValueIgnore,
	producers: []itemProducer{wordProducer},
	parseFn:   nil,
}

// ParserIP takes an IP item and produces a net.IP
var ParserIP = &ItemParser{
	valueType: ValueIP,
	producers: []itemProducer{ipProducer},
	parseFn:   parseIP,
}

func parseIP(input ...item) (value, error) {
	ip := net.ParseIP(input[0].val)
	return value{input, ip, ValueIP}, nil
}

func parseUser(input ...item) (value, error) {
	return value{input, input[0].val, ValueUser}, nil
}

// ParserRequest takes a quoted string item and produces an *http.Request
var ParserRequest = &ItemParser{
	valueType: ValueRequest,
	producers: []itemProducer{quotedStringProducer},
	parseFn:   parseRequest,
}

func parseRequest(input ...item) (value, error) {
	str := removeQuotes(input[0].val)
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nilVal(input), fmt.Errorf("invalid number of request parts")
	}

	req, err := http.NewRequest(parts[0], parts[1], nil)
	if err != nil {
		return nilVal(input), err
	}

	maj, min, ok := http.ParseHTTPVersion(parts[2])
	if !ok {
		return nilVal(input), fmt.Errorf("invalid HTTP version")
	}
	req.ProtoMajor, req.ProtoMinor = maj, min

	return value{input, req, ValueRequest}, nil
}

// ParserStatus takes an int item and produces an int64
var ParserStatus = &ItemParser{
	valueType: ValueStatus,
	producers: []itemProducer{intProducer},
	parseFn:   parseStatus,
}

func parseStatus(input ...item) (value, error) {
	status, err := strconv.ParseInt(input[0].val, 10, 64)
	if err != nil {
		return nilVal(input), err
	}
	return value{input, status, ValueStatus}, nil
}

// ParserBodyBytes takes an int item and produces an int64
var ParserBodyBytes = &ItemParser{
	valueType: ValueBodyBytes,
	producers: []itemProducer{intProducer},
	parseFn:   parseBodyBytes,
}

func parseBodyBytes(input ...item) (value, error) {
	status, err := strconv.ParseInt(input[0].val, 10, 64)
	if err != nil {
		return nilVal(input), err
	}
	return value{input, status, ValueBodyBytes}, nil
}

// ParserReferer takes a quoted string item and produces a *url.URL
var ParserReferer = &ItemParser{
	valueType: ValueReferer,
	producers: []itemProducer{quotedStringProducer},
	parseFn:   parseReferer,
}

func parseReferer(input ...item) (value, error) {
	str := removeQuotes(input[0].val)
	if str == "-" {
		return nilVal(input), nil
	}
	parsed, err := url.Parse(str)
	if err != nil {
		return nilVal(input), err
	}

	return value{input, parsed, ValueReferer}, nil
}

// ParserUser takes a word item and produces a string
var ParserUser = &ItemParser{
	valueType: ValueUser,
	producers: []itemProducer{wordProducer},
	parseFn:   parseUser,
}

// ParserUserAgent takes a quoted string item and produces a string
var ParserUserAgent = &ItemParser{
	valueType: ValueUserAgent,
	producers: []itemProducer{quotedStringProducer},
	parseFn:   parseUserAgent,
}

func parseUserAgent(input ...item) (value, error) {
	ua := removeQuotes(input[0].val)
	return value{input, ua, ValueUserAgent}, nil
}

func removeQuotes(input string) string {
	return strings.Trim(input, "\"`'")
}
