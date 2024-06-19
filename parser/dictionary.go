package parser

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// We use a struct to represent a dictionary entry instead of a string of only category
// This is to allow for future expansion of the dictionary entry
type DictEntry struct {
	Category string `json:"category" yaml:"category"`
}
type Dictionary map[string]DictEntry
type dictFileFormat int
type Wordcategory string
type ClassifiedWord struct {
	Word  string
	Class Wordcategory
}

// Enum of permissible dictionary file formats
const (
	JSON dictFileFormat = iota
	YAML
	UNKNOWN
)

const (
	NOUN         = "Noun"
	VERB         = "Verb"
	ADJECTIVE    = "Adjective"
	ADVERB       = "Adverb"
	PREPOSITION  = "Preposition"
	PRONOUN      = "Pronoun"
	DETERMINER   = "Determiner"
	CONJUNCTION  = "Conjunction"
	INTERJECTION = "Interjection"
)

var AvailableCategories = []Wordcategory{NOUN, VERB, ADJECTIVE, ADVERB, PREPOSITION, PRONOUN, DETERMINER, CONJUNCTION, INTERJECTION}

func NewDictionary() *Dictionary {
	return &Dictionary{}
}

func NewDictionaryFromFile(filename string) (*Dictionary, error) {
	var fileFormat = UNKNOWN
	if strings.HasSuffix(filename, ".json") {
		fileFormat = JSON
	} else if strings.HasSuffix(filename, ".yaml") {
		fileFormat = YAML
	} else {
		// Abort early if the file format is not recognized
		return nil, errors.New("invalid file format")
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	switch fileFormat {
	case JSON:
		return NewDictionaryFromJSONFile(file)
	case YAML:
		return NewDictionaryFromYAMLFile(file)
	}

	// This should never be reached
	return nil, errors.New("invalid file format")
}

func NewDictionaryFromJSONFile(f *os.File) (*Dictionary, error) {
	decoder := json.NewDecoder(f)
	dictionary := &Dictionary{}
	err := decoder.Decode(dictionary)
	if err != nil {
		return nil, err
	}

	return dictionary, nil
}

func NewDictionaryFromYAMLFile(f *os.File) (*Dictionary, error) {
	decoder := yaml.NewDecoder(f)
	dictionary := &Dictionary{}
	err := decoder.Decode(dictionary)
	if err != nil {
		return nil, err
	}

	return dictionary, nil
}

func (d *Dictionary) AddEntry(word string, category Wordcategory, smart bool) {
	(*d)[word] = DictEntry{Category: string(category)}

	// Some categories of the dictionary do more than just adding the word
	// For example nouns can be singular or plural, so we need to add both
	if smart {
		switch category {
		case NOUN:
			d.addNoun(word)
		case VERB:
			d.addVerb(word)
		}
	}
}

func (d *Dictionary) GetEntry(word string) (Wordcategory, bool) {
	entry, ok := (*d)[word]
	if !ok {
		return "", false
	}

	return Wordcategory(entry.Category), true
}

// Category functions

// This function will try to handle most cases of pluralizing and singularizing nouns
// It will add the word to the dictionary if it is not already present
// A best effort is made to handle most cases, but there are some edge cases that are not handled
// For example, "child" -> "children", "goose" -> "geese", "cactus" -> "cacti", "deer" -> "deer" are not handled
func (d *Dictionary) addNoun(noun string) {
	var toAdd string
	if strings.HasSuffix(noun, "ies") {
		toAdd = strings.TrimSuffix(noun, "ies") + "y"
	} else if strings.HasSuffix(noun, "s") {
		if strings.HasSuffix(noun, "es") {
			// Handles cases like "buses", "foxes"
			toAdd = strings.TrimSuffix(noun, "es")
		} else {
			// Handles cases like "bananas"
			toAdd = strings.TrimSuffix(noun, "s")
		}
	} else if strings.HasSuffix(noun, "y") {
		// Handles cases like "pony"
		toAdd = strings.TrimSuffix(noun, "y") + "ies"
	} else if strings.HasSuffix(noun, "f") {
		// Handles cases like "leaf"
		toAdd = strings.TrimSuffix(noun, "f") + "ves"
	} else if strings.HasSuffix(noun, "fe") {
		// Handles cases like "wife"
		toAdd = strings.TrimSuffix(noun, "fe") + "ves"
	} else if strings.HasSuffix(noun, "o") {
		// Handles cases like "potato"
		toAdd = noun + "es"
	} else {
		// Regular case
		toAdd = noun + "s"
	}

	// We want to make sure that the word is not already in the dictionary
	if _, ok := (*d)[toAdd]; !ok {
		(*d)[toAdd] = DictEntry{Category: "noun"}
	}
}

// This function will try to handle most cases of conjugating verbs
// Similar to the addNoun function, it will add the word to the dictionary if it is not already present
// Also similar to addNoun, a best effort is made, only that "best" is even worse in this case
func (d *Dictionary) addVerb(verb string) {
	var toAdd string
	if strings.HasSuffix(verb, "y") && !strings.HasSuffix(verb, "ay") && !strings.HasSuffix(verb, "ey") && !strings.HasSuffix(verb, "iy") && !strings.HasSuffix(verb, "oy") && !strings.HasSuffix(verb, "uy") {
		// Handles cases like "fly" -> "flies", but not "play" -> "plays"
		toAdd = strings.TrimSuffix(verb, "y") + "ies"
	} else if strings.HasSuffix(verb, "o") || strings.HasSuffix(verb, "sh") || strings.HasSuffix(verb, "ch") || strings.HasSuffix(verb, "s") || strings.HasSuffix(verb, "x") {
		// Handles cases like "go" -> "goes", "wash" -> "washes"
		toAdd = verb + "es"
	} else {
		// Regular case
		toAdd = verb + "s"
	}

	// We want to make sure that the word is not already in the dictionary
	if _, ok := (*d)[toAdd]; !ok {
		(*d)[toAdd] = DictEntry{Category: "verb"}
	}
}
