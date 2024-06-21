package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/QuisVenator/compi-tp/parser"
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
		for i := 0; i < 50; i++ {
			for _, word := range []Word{
				{Class: "Fruit", Word: "apple"},
				{Class: "Fruit", Word: "banana"},
				{Class: "Fruit", Word: "cherry"}} {
				classifiedWordsCh <- word
				time.Sleep(10 * time.Millisecond)
			}
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
	w := a.NewWindow("MNLTP")

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

	w.Resize(fyne.NewSize(1920, 1080))

	// Run update loop
	updateLoop := func() {
		startup(w)
		w.SetContent(content)

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
	}

	go updateLoop()
	w.ShowAndRun()
}

func startup(w fyne.Window) (p *parser.Parser, err error) {
	var exPath string
	ex, err := os.Executable()
	if err == nil {
		exPath = filepath.Dir(ex)
	} else {
		exPath = "./"
	}

	inputFilePaths := []string{}
	dictPath := exPath + "dictionary.json"
	outputPath := exPath + "output.csv"
	startCh := make(chan struct{})

	// Widgets
	dictionarySelectedLabel := container.NewHBox(
		widget.NewIcon(theme.DocumentIcon()),
		widget.NewLabel(dictPath),
	)
	outputLabel := widget.NewLabel(outputPath)
	inputFileList := widget.NewList(
		func() int {
			return len(inputFilePaths)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("input.txt"),
			)
		},
		func(i widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(inputFilePaths[i])
		},
	)
	inputFileList.Resize(fyne.NewSize(inputFileList.MinSize().Width, inputFileList.MinSize().Height*3))

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Welcome to MNLTP!"),
			container.NewVBox(
				widget.NewLabel("Please select dictionary to use:"),
				dictionarySelectedLabel,
				widget.NewButton("Open Dictionary", func() {
					d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						if reader == nil {
							dictPath = "dictionary.json"
						} else {
							dictPath = reader.URI().Path()
							defer reader.Close()
						}
						dictionarySelectedLabel.Objects[1].(*widget.Label).SetText(dictPath)
						dictionarySelectedLabel.Refresh()
					}, w)
					d.SetFilter(storage.NewExtensionFileFilter([]string{".json", ".yaml"}))
					d.Show()
				}),
			),
		),
		container.NewBorder(
			container.NewHBox(
				widget.NewLabel("Save output to:"),
				outputLabel,
				widget.NewButton("Change", func() {
					dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						if writer != nil {
							outputPath = writer.URI().Path()
							defer writer.Close()
						}
						outputLabel.SetText(outputPath)
					}, w)
				}),
			),
			widget.NewButton("Start", func() {
				startCh <- struct{}{}
			}),
			nil,
			nil,
			nil,
		),
		nil,
		nil,
		container.NewBorder(
			widget.NewLabel("Please select input files:"),
			widget.NewButton("Add Input File", func() {
				dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					if reader == nil {
						return
					}
					defer reader.Close()

					inputFilePaths = append(inputFilePaths, reader.URI().Path())
					inputFileList.Refresh()
				}, w)
			}),
			nil,
			nil,
			inputFileList,
		),
	)

	w.SetContent(content)

	// Run loop for startup
	for {
		<-startCh
		p, err := parser.NewParser(dictPath, inputFilePaths, outputPath, nil)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			return p, nil
		}
	}
}
