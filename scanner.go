package main

import (
	"log"
	"strings"
	"unicode"
	"unicode/utf8"
)

// this code is heavily borrowed from text/template
// https://github.com/golang/go/blob/07b8011393dc3d3a78b3cd0857a31da339985994/src/text/template/parse/lex.go

const (
	itemError itemType = iota
	itemEOF
	itemSpace
	itemNewline
	itemNumber
	itemInt
	itemWord
	itemString
	itemLeftDelimiter
	itemRightDelimiter
	itemQuote
	itemQuotedString
	itemIP
	itemDash
)

const (
	digits = "0987654321"
	eof    = rune(0)
)

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

var ipProducer = itemProducer{scanIP, itemIP}
var spaceProducer = itemProducer{scanSpace, itemSpace}
var wordProducer = itemProducer{scanWord, itemWord}
var numberProducer = itemProducer{scanNumber, itemNumber}
var intProducer = itemProducer{scanInt, itemInt}
var leftDelimProducer = itemProducer{scanLeftDelimiter, itemLeftDelimiter}
var rightDelimProducer = itemProducer{scanRightDelimiter, itemRightDelimiter}
var quotedStringProducer = itemProducer{scanQuotedString, itemQuotedString}

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

// next returns the next rune in the input.
func (s *Scanner) next() rune {
	if int(s.pos) >= len(s.input) {
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
	s.items <- item{t, s.start, s.input[s.start:s.pos]}
	s.start = s.pos
}

// ignore skips over the pending input before this point.
func (s *Scanner) ignore() {
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

// acceptSequence consumes a string if found & returns true, false if not
func (s *Scanner) acceptSequence(valid string) bool {
	if strings.HasPrefix(s.input[s.pos:], valid) {
		s.pos += len(valid)
		return true
	}
	return false
}

// acceptUntil consumes to 'end' or eof; returns true if it accepts, false otherwise
func (s *Scanner) acceptUntil(end rune) bool {
	if s.peek() == end || s.peek() == eof {
		return false
	}
	for r := s.next(); r != end && r != eof; r = s.next() {
	}
	s.backup()
	return true
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

// lineNumber reports what line we're on, based on the position of the previous
// item returned by nextItem; this way we don't have to worry about peek double
// counting
func (s *Scanner) lineNumber() int {
	return 1 + strings.Count(s.input[:s.lastPos], "\n")
}

// nextItem returns the next item from the input; called by parser
func (s *Scanner) nextItem() item {
	item := <-s.items
	s.lastPos = item.pos
	return item
}

// drain runs through output so lexing goroutine exists; called by parser
func (s *Scanner) drain() {
	for range s.items {
	}
}

// run steps through the state machine
func (s *Scanner) run() {
	for _, producer := range s.itemOrder {
		s.state = producer.fn(s)
		// follow any returned stateFns until nil
		for s.state != nil {
			s.state = s.state(s)
		}
		// need to scan spaces in between items
		// TODO: make auto-space-scan configurable?
		s.state = scanSpace(s)
	}
	close(s.items)
}

func scanLine(s *Scanner) stateFn {
	r := s.next()
	if r == eof {
		s.emit(itemEOF)
		return nil
	} else if isNewline(r) {
		return scanNewline
	}

	log.Fatalf("WHAT IS '%c'", r)
	return nil
}

// scanSpace scans a run of space characters; one space already seen
func scanSpace(s *Scanner) stateFn {
	for isSpace(s.peek()) {
		s.next()
	}
	s.emit(itemSpace)
	return scanLine
}

func scanNewline(s *Scanner) stateFn {
	for isNewline(s.peek()) {
		s.next()
	}
	s.emit(itemNewline)
	return scanLine
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

// isSpace returns true if space or tab
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// UNUSED?
func isNewline(r rune) bool {
	return r == '\r' || r == '\n'
}

func isGenericDelim(r rune) bool {
	return isSpace(r) || isLeftDelim(r) || isRightDelim(r)
}

func isLeftDelim(r rune) bool {
	return strings.ContainsRune("([{<", r)
}

func isRightDelim(r rune) bool {
	return strings.ContainsRune(")]}>", r)
}

func isQuote(r rune) bool {
	return r == '\'' || r == '"'
}

func scan(input string, itemOrder []itemProducer) *Scanner {
	s := &Scanner{
		input:     input,
		items:     make(chan item),
		itemOrder: itemOrder,
	}
	go s.run()
	return s
}
