package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var whitespaceOrPunctuation = regexp.MustCompile(`[\s\p{P}]+`)

type parser struct {
	dict     *Dictionary
	dictFile string
	input    []*os.File
	output   *os.File
	inpath   []string
	outpath  string
	Outchan  chan ClassifiedWord
	Newword  chan string
	classes  <-chan Wordcategory
}

func NewParser(dict string, inpath []string, outpath string, classchan <-chan Wordcategory) (*parser, error) {
	dictionary, err := NewDictionaryFromFile(dict)
	if err != nil {
		return nil, err
	}
	var p parser
	p.dictFile = dict
	p.dict = dictionary
	p.inpath = inpath
	p.outpath = outpath

	for _, path := range inpath {
		input, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		p.input = append(p.input, input)
	}
	p.output, err = os.Create(outpath)
	if err != nil {
		return nil, err
	}

	p.Outchan = make(chan ClassifiedWord)
	p.Newword = make(chan string)
	p.classes = classchan

	return &p, nil
}

func (p *parser) Parse() error {
	positions := make(map[Wordcategory]string)
	words := make(map[Wordcategory]string)

	for i, input := range p.input {
		wordnum := 0
		Scanner := bufio.NewScanner(input)
		Scanner.Split(SplitWords)

		for Scanner.Scan() {
			word := Scanner.Text()
			if word == "" {
				continue
			}
			word = strings.ToLower(word)
			wordnum++
			class, ok := p.dict.GetEntry(word)
			if !ok {
				p.Newword <- word
				class = <-p.classes
				p.dict.AddEntry(word, class, false)
			}
			p.Outchan <- ClassifiedWord{word, class}
			positions[class] += fmt.Sprintf("TXT#%d-%d,", i, wordnum)
			words[class] += word + ","
		}
	}

	p.output.WriteString("TOKEN,LEXEMAS,POSICIONES\n")
	for _, cat := range AvailableCategories {
		p.output.WriteString(string(cat))
		p.output.WriteString(";")
		p.output.WriteString(strings.TrimSuffix(words[cat], ","))
		p.output.WriteString(";")
		p.output.WriteString(strings.TrimSuffix(positions[cat], ","))
		p.output.WriteString("\n")
	}

	p.Outchan <- ClassifiedWord{"", EOF}
	return nil
}

func (p *parser) Close() {
	for _, input := range p.input {
		input.Close()
	}
	p.output.Close()
	close(p.Outchan)
	close(p.Newword)
	err := p.dict.SaveToFile(p.dictFile)
	if err != nil {
		fmt.Println(err)
	}
}

func SplitWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if match := whitespaceOrPunctuation.FindIndex(data); match != nil {
		return match[1], data[:match[0]], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}
