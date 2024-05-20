package parser

import (
	"encoding/json"
	"os"
)

type DFA struct {
	Transitions  map[string]map[string]string `json:"transitions" yaml:"transitions"`
	CurrentState string                       `json:"currentState" yaml:"currentState"`
	AcceptStates map[string]bool              `json:"acceptStates" yaml:"acceptStates"`
}

func NewStartingDFA() (*DFA, error) {
	return NewDfaFromFile("configs/english_starting_dfa.json")
}

func NewDfaFromFile(filename string) (*DFA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	dfa := &DFA{}
	err = decoder.Decode(dfa)
	if err != nil {
		return nil, err
	}

	return dfa, nil
}

func (dfa *DFA) Transition(input string) bool {
	nextState, ok := dfa.Transitions[dfa.CurrentState][input]
	if !ok {
		return false
	}
	dfa.CurrentState = nextState
	return true
}
