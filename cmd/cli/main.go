package main

import (
	"fmt"
	"os"

	"github.com/QuisVenator/compi-tp/parser"
)

func usage() string {
	return `
	Usage: parser <input.txt> [output.csv [dictionary]]
	The program reads an input file to analyze. Optionally the output filename can be specified (default: output.csv) and the dictionary file (default: dictionary.json).
	The dictionary must be in either JSON or YAML format.
	`
}

func main() {
	// Parse the command line arguments
	if len(os.Args) < 2 {
		fmt.Println(usage())
		return
	} else if len(os.Args) > 4 {
		fmt.Println("Too many arguments")
		fmt.Println(usage())
		return
	}

	inputFilename := os.Args[1]
	outputFilename := "output.csv"
	dictionaryFilename := "dictionary.json"
	if len(os.Args) > 2 {
		outputFilename = os.Args[2]
	}
	if len(os.Args) > 3 {
		dictionaryFilename = os.Args[3]
	}

	// Create the parser
	classchan := make(chan parser.Wordcategory)
	p, err := parser.NewParser(dictionaryFilename, []string{inputFilename}, outputFilename, classchan)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer p.Close()

	// Start the parser
	go p.Parse()

	// Start the display
	displayResults(p.Outchan, p.Newword, classchan)
}

// This function reads words and their classification from the output channel and displays them on the console.
// Additionally, it reads words from the newwords channel and asks the use to classify them. The classification is then sent on the classes channel.
func displayResults(output <-chan parser.ClassifiedWord, newwords <-chan string, classes chan<- parser.Wordcategory) {
	for {
		select {
		case word := <-output:
			if word.Class == parser.EOF {
				return
			}
			fmt.Printf("%s: %s\n", word.Word, word.Class)
		case newword := <-newwords:
			fmt.Printf("Please classify the word '%s':\n", newword)
			for class := range parser.AvailableCategories {
				fmt.Printf("%d: %s\n", class+1, parser.AvailableCategories[class])
			}
			var class int
			fmt.Scan(&class)
			for class < 1 || class > len(parser.AvailableCategories) {
				fmt.Println("Invalid class. Please enter a valid class:")
				fmt.Scan(&class)
			}
			classes <- parser.AvailableCategories[class-1]
		}
	}
}
