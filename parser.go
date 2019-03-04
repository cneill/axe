package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type LogLine struct {
	IP          string
	User        string
	Time        time.Time
	Method      string
	Path        string
	HTTPVersion string
	// Request string
	Status    int64
	BodyBytes int64
	Referer   string
	UserAgent string
}

type Parser struct {
	s   *Scanner
	buf struct {
		item item
		raw  string
		n    int
	}
	itemOrder []itemProducer
}

// $ip - $user [$time $tz] "$req" $status $bytes "$ref" "$ua"
var itemOrder = []itemProducer{
	ipProducer,           // X 0: $ip
	wordProducer,         // X 1: -
	wordProducer,         // X 2: $user
	leftDelimProducer,    // X 3: [
	wordProducer,         // X 4: $time
	wordProducer,         // X 5: $tz
	rightDelimProducer,   // X 6: ]
	quotedStringProducer, // 7: "$req"
	intProducer,          // X 8: $status
	intProducer,          // X 9: $bytes
	quotedStringProducer, // X 10: "$ref"
	quotedStringProducer, // X 11: "$ua"
}

func NewParser(input string) *Parser {
	p := &Parser{
		itemOrder: itemOrder,
	}

	p.s = scan(input, p.itemOrder)
	return p
}

// TODO: MAKE THIS CONFIGURABLE
func (p *Parser) ParseLine() (*LogLine, error) {
	var items []item
	for _, producer := range p.itemOrder {
		item, raw := p.scanIgnoreSpaces()

		if item.typ != producer.typ {
			// log.Fatalf("GOT UNEXPECTED VALUE: %v (%s)\nEXPECTED: %v", item.typ, raw, producer.typ)
			fmt.Printf("GOT UNEXPECTED VALUE: %v (%s)\nEXPECTED: %v\n", item.typ, raw, producer.typ)
			return nil, nil
		}

		// fmt.Printf("GOT '%s'\n", raw)
		items = append(items, item)
	}

	fullTime := items[4].val + " " + items[5].val
	parsedTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", fullTime)
	if err != nil {
		return nil, err
	}

	status, err := strconv.ParseInt(items[8].val, 10, 64)
	if err != nil {
		return nil, err
	}

	bytes, err := strconv.ParseInt(items[9].val, 10, 64)
	if err != nil {
		return nil, err
	}

	ll := &LogLine{
		IP:        items[0].val,
		User:      items[2].val,
		Time:      parsedTime,
		Status:    status,
		BodyBytes: bytes,
		Referer:   strings.Trim(items[10].val, "\""),
		UserAgent: strings.Trim(items[11].val, "\""),
	}

	return ll, nil
}

func (p *Parser) scan() (item, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.item, p.buf.raw
	}

	item := p.s.nextItem()

	p.buf.item, p.buf.raw = item, item.val
	return p.buf.item, p.buf.raw
}

func (p *Parser) scanIgnoreSpaces() (item, string) {
	item, raw := p.scan()
	for item.typ == itemSpace {
		item, raw = p.scan()
	}
	return item, raw
}

func (p *Parser) unscan() {
	p.buf.n = 1
}
