package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// TODO: don't bother parsing bits that aren't relevant - if printing IPs, just parse them and leave the rest nil

var axeWG = &sync.WaitGroup{}
var numLinesMutex = &sync.RWMutex{}

// Axe controls parsing of STDIN
type Axe struct {
	numWorkers int
	source     *os.File
	printFunc  func(*LogLine)
	errFunc    func(error)

	inChan   chan string
	outChan  chan *LogLine
	errChan  chan error
	numLines int
}

type llFunc func(*LogLine)
type errFunc func(error)

// NewAxe returns a prepared *Axe
func NewAxe(numWorkers int, pf llFunc, ef errFunc) *Axe {
	a := &Axe{
		numWorkers: numWorkers,
		source:     os.Stdin,
		printFunc:  pf,
		errFunc:    ef,

		inChan:  make(chan string, 1024),
		outChan: make(chan *LogLine, 1024),
		errChan: make(chan error),
	}

	return a
}

// Start kicks off the workers, reads the source, and displays the output
func (a *Axe) Start() {
	// start our readWorker to read raw strings from STDIN
	axeWG.Add(1)
	go a.readWorker(axeWG.Done)

	// start our outWorker to start printing lines
	axeWG.Add(1)
	go a.outWorker(axeWG.Done)

	// start our inWorkers to start parsing raw lines
	for i := 0; i < a.numWorkers; i++ {
		axeWG.Add(1)
		go a.inWorker(axeWG.Done)
	}

	axeWG.Wait()
}

// readWorker reads raw strings from STDIN
func (a *Axe) readWorker(done func()) {
	defer done()
	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		text := s.Text()
		a.inChan <- text
	}
	close(a.inChan)
}

func (a *Axe) incrNumLines() {
	numLinesMutex.Lock()
	a.numLines++
	numLinesMutex.Unlock()
}

// inWorker parses strings from readWorker into LogLines
func (a *Axe) inWorker(done func()) {
	defer done()
	parser := NewParser(nginxItemOrder)
	for input := range a.inChan {
		a.incrNumLines()
		ll, err := parser.ParseLine(input)
		if err != nil {
			a.errChan <- fmt.Errorf("%d:%v", a.numLines, err)
		} else {
			a.outChan <- ll
		}
	}
	close(a.outChan)
}

// outWorker prints LogLines with a.printFunc
func (a *Axe) outWorker(done func()) {
	defer func() {
		close(a.errChan)
		done()
	}()
	for {
		select {
		case output := <-a.outChan:
			if output == nil {
				return
			}
			a.printFunc(output)
		case err := <-a.errChan:
			a.errFunc(err)
		}
	}
}
