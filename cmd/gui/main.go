package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Word struct {
	Class string
	Word  string
}

func main() {
	classifiedWordsCh := make(chan Word)
	newWordsCh := make(chan string)
	userChosenClassCh := make(chan Word)

	// Simulating incoming words for testing purposes
	go func() {
		for _, word := range []string{"apple", "banana", "cherry"} {
			newWordsCh <- word
			word := <-userChosenClassCh
			classifiedWordsCh <- word
		}
		// List of 50 already classified words
		for _, word := range []Word{
			{Class: "Fruit", Word: "apple"},
			{Class: "Fruit", Word: "banana"},
			{Class: "Fruit", Word: "cherry"}} {
			classifiedWordsCh <- word
			time.Sleep(10 * time.Millisecond)
		}

	}()

	runGui(classifiedWordsCh, newWordsCh, userChosenClassCh)
}

func runGui(classifiedWordsCh chan Word, newWordsCh chan string, userChosenClassCh chan Word) {
	processedWords := make(map[string]struct{})
	processedWordsText := ""
	classifiedText := ""
	newWordsCount := 0
	processedWordsCount := 0

	a := app.New()
	w := a.NewWindow("Word Classifier")

	// Widgets
	header := widget.NewRichTextFromMarkdown("# MNLTP\n___")
	header.Wrapping = fyne.TextWrapWord
	header.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter

	processedWordsTextField := widget.NewRichTextFromMarkdown("# Processed Words\n")
	processedWordsTextField.Wrapping = fyne.TextWrapWord
	processedWordsTextField.Segments = append(processedWordsTextField.Segments, &widget.TextSegment{Text: processedWordsText})
	classifiedTextField := widget.NewRichTextFromMarkdown("# Classified Words\n")
	classifiedTextField.Wrapping = fyne.TextWrapWord
	classifiedTextField.Segments = append(classifiedTextField.Segments, &widget.TextSegment{Text: classifiedText})

	infoLabel := widget.NewLabelWithStyle("Info", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	processedWordsCountLabel := widget.NewLabel(fmt.Sprintf("Total Words Processed: %d", processedWordsCount))
	distinctWordsLabel := widget.NewLabel(fmt.Sprintf("Distinct Words: %d", len(processedWords)))
	newWordsCountLabel := widget.NewLabel(fmt.Sprintf("New Words Classified: %d", newWordsCount))

	// Layout
	content := container.NewVBox(
		header,
		container.New(
			layout.NewGridLayout(2),
			processedWordsTextField,
			classifiedTextField,
		),
		layout.NewSpacer(),
		infoLabel,
		processedWordsCountLabel,
		distinctWordsLabel,
		newWordsCountLabel,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(1920, 1080))

	// Run update loop
	go func() {
		for {
			select {
			case word := <-classifiedWordsCh:
				processedWords[word.Word] = struct{}{}
				processedWordsText += word.Word + " "
				classifiedText += word.Class + " "
				processedWordsCount++
				processedWordsTextField.Segments[1].(*widget.TextSegment).Text = processedWordsText
				processedWordsTextField.Refresh()
				classifiedTextField.Segments[1].(*widget.TextSegment).Text = classifiedText
				classifiedTextField.Refresh()
				processedWordsCountLabel.SetText(fmt.Sprintf("Total Words Processed: %d", processedWordsCount))
				distinctWordsLabel.SetText(fmt.Sprintf("Distinct Words: %d", len(processedWords)))
			case word := <-newWordsCh:
				var dia *dialog.CustomDialog

				// Create classification dialog
				classes := []string{"Fruit", "Vegetable", "Animal"}
				prompt := widget.NewLabel(fmt.Sprintf("Classify '%s'", word))

				classBtns := make([]fyne.CanvasObject, len(classes))
				for i, class := range classes {
					classBtns[i] = widget.NewButton(class, func() {
						userChosenClassCh <- Word{Class: class, Word: word}
						newWordsCount++
						newWordsCountLabel.SetText(fmt.Sprintf("New Words Classified: %d", newWordsCount))
						dia.Hide()
					})
				}

				diaContent := container.NewVBox(
					append([]fyne.CanvasObject{prompt}, classBtns...)...,
				)

				dia = dialog.NewCustomWithoutButtons("Classify Word", diaContent, w)
				dia.Show()
			}
		}
	}()

	w.ShowAndRun()
}
