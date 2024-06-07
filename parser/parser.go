package parser

import (
	"regexp"
)

var whitespaceOrPunctuation = regexp.MustCompile(`[\s\p{P}]`)
