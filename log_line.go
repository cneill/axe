package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// LogLine represents a parsed line from a log
type LogLine struct {
	IP        net.IP
	Hostname  string // UNUSED - for --resolve flag
	User      string
	Time      time.Time
	Request   *http.Request
	Status    int64
	BodyBytes int64
	Referer   *url.URL
	UserAgent string

	Error error
}

// TODO: make this customizable
// $ip - $user [$time $tz] "$req" $status $bytes "$ref" "$ua"
func (l *LogLine) String() string {
	var method, path string
	var ip, referer = "-", "-"
	var ver = "HTTP/1.1"
	if l.IP != nil {
		ip = l.IP.String()
	}
	if l.Request != nil {
		req := l.Request
		method = req.Method
		path = req.URL.String()
		ver = req.Proto
	}
	if l.Referer != nil {
		referer = l.Referer.String()
	}
	return fmt.Sprintf(
		"%s - %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"",
		ip, l.User, l.Time.Format(nginxTimeFormat), method,
		path, ver, l.Status, l.BodyBytes, referer, l.UserAgent,
	)
}

func (l *LogLine) invalidValueErr(ok bool, input value) bool {
	if !ok {
		l.Error = fmt.Errorf("invalid %s: %v", input.valueType, input.obj)
		return true
	}
	return false
}

func (l *LogLine) add(input value) *LogLine {
	switch input.valueType {
	case ValueBodyBytes:
		bodyBytes, ok := input.obj.(int64)
		if !l.invalidValueErr(ok, input) {
			l.BodyBytes = bodyBytes
		}
	case ValueIP:
		ip, ok := input.obj.(net.IP)
		if !l.invalidValueErr(ok, input) {
			l.IP = ip
		}
	case ValueReferer:
		u, ok := input.obj.(*url.URL)
		if !l.invalidValueErr(ok, input) {
			l.Referer = u
		}
	case ValueRequest:
		req, ok := input.obj.(*http.Request)
		if !l.invalidValueErr(ok, input) {
			l.Request = req
		}
	case ValueStatus:
		status, ok := input.obj.(int64)
		if !l.invalidValueErr(ok, input) {
			l.Status = status
		}
	case ValueTime:
		t, ok := input.obj.(time.Time)
		if !l.invalidValueErr(ok, input) {
			l.Time = t
		}
	case ValueUser:
		user, ok := input.obj.(string)
		if !l.invalidValueErr(ok, input) {
			l.User = user
		}
	case ValueUserAgent:
		ua, ok := input.obj.(string)
		if !l.invalidValueErr(ok, input) {
			l.UserAgent = ua
		}
	case ValueIgnore:
	case ValueNil:
	default:
		l.Error = fmt.Errorf("invalid item: %v", input)
	}
	return l
}
