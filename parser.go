package main

// this borrows from Ben Johnson's tutorial on parsers: https://blog.gopheracademy.com/advent-2014/parsers-lexers/
// its license is below
/*
The MIT License (MIT)

Copyright (c) 2014 Ben Johnson

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

// Parser handles collecting items and parsing them into their final values
type Parser struct {
	s   *Scanner
	buf struct {
		item item
		raw  string
		n    int
	}
	itemOrder []*ItemParser
}

// NewParser returns a prepared *Parser with an attached *Scanner
func NewParser(itemOrder []*ItemParser) *Parser {
	var producerOrder = []itemProducer{}

	for _, order := range itemOrder {
		producerOrder = append(producerOrder, order.producers...)
	}

	return &Parser{
		itemOrder: itemOrder,
		s:         NewScanner(producerOrder),
	}
}

// ParseLine takes a raw string line as input and returns a *LogLine, or error
// TODO: MAKE THIS CONFIGURABLE
func (p *Parser) ParseLine(input string) (*LogLine, error) {
	var ll = &LogLine{}
	defer p.reset()

	go p.s.run(input)

	for _, ip := range p.itemOrder {
		var items []item

		// gather all our items, make sure we get the right types
		for _, producer := range ip.producers {
			it, _ := p.scanIgnoreSpaces()
			if it.typ == producer.typ {
				items = append(items, it)
			} else {
				items = append(items, item{itemError, it.pos, it.val})
			}
		}

		// get the value for these items
		val, err := ip.parse(items...)
		if err != nil {
			ll.Error = err
		}

		ll.add(val)
	}

	return ll, ll.Error
}

func (p *Parser) scan() (item, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.item, p.buf.raw
	}

	it := p.s.nextItem()

	p.buf.item, p.buf.raw = it, it.val
	return p.buf.item, p.buf.raw
}

func (p *Parser) scanIgnoreSpaces() (item, string) {
	it, raw := p.scan()
	if it.typ == itemSpace {
		it, raw = p.scan()
	}
	return it, raw
}

func (p *Parser) reset() {
	p.buf.item = item{}
	p.buf.raw = ""
	p.buf.n = 0
	p.s.reset()
}

// UNUSED
/*
func (p *Parser) unscan() {
	p.buf.n = 1
}
*/
