package main

import (
	"strings"
	"unicode/utf8"
)

// this code is heavily based on text/template
// https://github.com/golang/go/blob/07b8011393dc3d3a78b3cd0857a31da339985994/src/text/template/parse/lex.go

type stateFn func(*Scanner) stateFn
type runeFn func(rune) bool
type itemProducer struct {
	fn  stateFn
	typ itemType
}
type itemType int
type item struct {
	typ itemType
	pos int
	val string
}

// Scanner handles collecting individual pieces of a log line
type Scanner struct {
	input   string
	state   stateFn
	pos     int
	start   int
	width   int
	lastPos int
	items   chan item

	itemOrder []itemProducer
}

// NewScanner returns a *Scanner for user by a *Parser
func NewScanner(order []itemProducer) *Scanner {
	s := &Scanner{
		items:     make(chan item),
		itemOrder: order,
	}
	return s
}

// run steps through the state machine
func (s *Scanner) run(input string) {
	s.input = input
	for i, producer := range s.itemOrder {
		s.state = producer.fn(s)
		// follow any returned stateFns until nil
		for s.state != nil {
			s.state = s.state(s)
		}
		// need to scan spaces in between items
		// TODO: make auto-space-scan configurable?
		if i != len(s.itemOrder)-1 {
			s.state = scanSpace(s)
		}
	}
}

func (s *Scanner) reset() {
	close(s.items)
	s.input = ""
	s.state = nil
	s.pos = 0
	s.start = 0
	s.width = 0
	s.lastPos = 0
	s.items = make(chan item)
}

// next returns the next rune in the input.
func (s *Scanner) next() rune {
	if s.pos >= len(s.input) {
		s.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(s.input[s.pos:])
	s.width = w
	s.pos += s.width
	return r
}

// backup steps back one rone. Can only be called once per call of next.
func (s *Scanner) backup() {
	s.pos -= s.width
}

// peek returns but does not consume the next rune in the input.
func (s *Scanner) peek() rune {
	r := s.next()
	s.backup()
	return r
}

// emit passes an item to the items channel.
func (s *Scanner) emit(t itemType) {
	i := item{t, s.start, s.input[s.start:s.pos]}
	s.items <- i
	s.start = s.pos
}

// accept consumes the next rune if it's from the valid set
func (s *Scanner) accept(valid string) bool {
	if strings.ContainsRune(valid, s.next()) {
		return true
	}
	s.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set
func (s *Scanner) acceptRun(valid string) {
	for strings.ContainsRune(valid, s.next()) {
	}
	s.backup()
}

// acceptUntilRuneFn consumes until 'end' returns true
func (s *Scanner) acceptUntilRuneFn(end runeFn) bool {
	accepted := false
	for r := s.peek(); !end(r) && r != eof; r = s.peek() {
		s.next()
		accepted = true
	}
	return accepted
}

// nextItem returns the next item from the input; called by parser
func (s *Scanner) nextItem() item {
	item := <-s.items
	s.lastPos = item.pos
	return item
}

// UNUSED
// acceptSequence consumes a string if found & returns true, false if not
/*
func (s *Scanner) acceptSequence(valid string) bool {
	if strings.HasPrefix(s.input[s.pos:], valid) {
		s.pos += len(valid)
		return true
	}
	return false
}
*/

// acceptUntil consumes to 'end' or eof; returns true if it accepts, false otherwise
/*
func (s *Scanner) acceptUntil(end rune) bool {
	if s.peek() == end || s.peek() == eof {
		return false
	}
	for r := s.next(); r != end && r != eof; r = s.next() {
	}
	s.backup()
	return true
}
*/

// drain runs through output so lexing goroutine exists; called by parser
/*
func (s *Scanner) drain() {
	for range s.items {
	}
}
*/

// ignore skips over the pending input before this point. - UNUSED FOR NOW
/*
func (s *Scanner) ignore() {
	s.start = s.pos
}
*/

// scanSpace scans a run of space characters; one space already seen
func scanSpace(s *Scanner) stateFn {
	for isSpace(s.peek()) {
		s.next()
	}
	s.emit(itemSpace)
	return nil
}

// here we are merely splitting on space - not dealing with quotes
func scanWord(s *Scanner) stateFn {
	s.acceptUntilRuneFn(isGenericDelim)
	s.emit(itemWord)
	return nil
}

func scanInt(s *Scanner) stateFn {
	s.acceptRun(digits)
	s.emit(itemInt)
	return nil
}

/*
func scanNumber(s *Scanner) stateFn {
	if s.accept("-") {
		// we have either a "-" or numeric characters
		if !unicode.IsNumber(s.peek()) {
			s.emit(itemError)
			return nil
		}
	}
	s.acceptRun(digits)
	s.accept(".")
	s.acceptRun(digits)
	s.emit(itemNumber)
	return nil
}
*/

func scanIP(s *Scanner) stateFn {
	s.acceptRun(digits)
	s.accept(".")
	s.acceptRun(digits)
	s.accept(".")
	s.acceptRun(digits)
	s.accept(".")
	s.acceptRun(digits)
	s.emit(itemIP)
	return nil
}

func scanLeftDelimiter(s *Scanner) stateFn {
	r := s.next()
	if isLeftDelim(r) {
		s.emit(itemLeftDelimiter)
		return nil
	}
	s.emit(itemError)
	return nil
}

func scanRightDelimiter(s *Scanner) stateFn {
	r := s.next()
	if isRightDelim(r) {
		s.emit(itemRightDelimiter)
		return nil
	}
	s.emit(itemError)
	return nil
}

func scanQuotedString(s *Scanner) stateFn {
	r := s.next()
	if !isQuote(r) {
		s.emit(itemError)
		return nil
	}
	s.acceptUntilRuneFn(isQuote)
	r = s.next()
	if !isQuote(r) {
		s.emit(itemError)
		return nil
	}
	s.emit(itemQuotedString)
	return nil
}

// isSpace returns true if r is space or tab
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isGenericDelim returns true if r is space/left delim/right delim
func isGenericDelim(r rune) bool {
	return isSpace(r) || isLeftDelim(r) || isRightDelim(r)
}

// isLeftDelim returns true if r is one of ( [ { <
func isLeftDelim(r rune) bool {
	return strings.ContainsRune("([{<", r)
}

// isRightDelim returns true if r is one of ) ] } >
func isRightDelim(r rune) bool {
	return strings.ContainsRune(")]}>", r)
}

// isQuote returns true if r is one of " ' `
func isQuote(r rune) bool {
	return strings.ContainsRune("\"'`", r)
}

// UNUSED
/*
func isNewline(r rune) bool {
	return r == '\r' || r == '\n'
}
*/

var intProducer = itemProducer{scanInt, itemInt}
var ipProducer = itemProducer{scanIP, itemIP}
var leftDelimProducer = itemProducer{scanLeftDelimiter, itemLeftDelimiter}
var rightDelimProducer = itemProducer{scanRightDelimiter, itemRightDelimiter}
var quotedStringProducer = itemProducer{scanQuotedString, itemQuotedString}
var wordProducer = itemProducer{scanWord, itemWord}

// UNUSED
// var spaceProducer = itemProducer{scanSpace, itemSpace}
// var numberProducer = itemProducer{scanNumber, itemNumber}
