package main

import (
	"fmt"
	"os"

	"github.com/QuisVenator/compi-tp/tokenizer"
)

func usage() string {
	return `
	Usage: tokenizer <input.txt> [output.csv [dictionary]]
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

	// Create the tokenizer
	classchan := make(chan tokenizer.Wordcategory)
	infochan := make(chan tokenizer.Runinfo)
	p, err := tokenizer.NewTokenizer(dictionaryFilename, []string{inputFilename}, outputFilename, classchan, infochan)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer p.Close()

	// Start the tokenizer
	go p.Parse()

	// Start the display
	displayResults(p.Outchan, p.Newword, classchan, infochan)
}

// This function reads words and their classification from the output channel and displays them on the console.
// Additionally, it reads words from the newwords channel and asks the use to classify them. The classification is then sent on the classes channel.
func displayResults(output <-chan tokenizer.ClassifiedWord, newwords <-chan string, classes chan<- tokenizer.Wordcategory, infochan chan tokenizer.Runinfo) {
	for {
		select {
		case word := <-output:
			if word.Class == tokenizer.EOF {
				info := <-infochan
				fmt.Printf("Word count: %d\n", info.WordCount)
				fmt.Printf("Distinct word count: %d\n", info.DistinctWordCount)
				for cat, count := range info.WordPerCategory {
					fmt.Printf("Word count for category %s: %d\n", cat, count)
				}
				for cat, count := range info.DistinctWordPerCategory {
					fmt.Printf("Distinct word count for category %s: %d\n", cat, count)
				}
				fmt.Printf("New word count: %d\n", info.NewWordCount)
				for cat, count := range info.NewWordPerCategory {
					fmt.Printf("New word count for category %s: %d\n", cat, count)
				}
				fmt.Printf("File count: %d\n", info.FileCount)
				fmt.Printf("Time spent: %s\n", info.TimeSpent)
				fmt.Printf("Time waited: %s\n", info.TimeWaited)
				close(infochan)
				return
			}
			fmt.Printf("%s: %s\n", word.Word, word.Class)
		case newword := <-newwords:
			fmt.Printf("Please classify the word '%s':\n", newword)
			for class := range tokenizer.AvailableCategories {
				fmt.Printf("%d: %s\n", class+1, tokenizer.AvailableCategories[class])
			}
			var class int
			fmt.Scan(&class)
			for class < 1 || class > len(tokenizer.AvailableCategories) {
				fmt.Println("Invalid class. Please enter a valid class:")
				fmt.Scan(&class)
			}
			classes <- tokenizer.AvailableCategories[class-1]
		}
	}
}
