package main

import (
	"fmt"
	"os"

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

func main() {
	runGui()
}

func runGui() {
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
		userChosenClassCh := make(chan parser.Wordcategory)
		p := startup(w, userChosenClassCh)
		classifiedWordsCh := p.Outchan
		newWordsCh := p.Newword
		w.SetContent(content)
		go p.Parse()

		for {
			select {
			case word := <-classifiedWordsCh:
				processedWords[word.Word] = struct{}{}
				processedWordsText += word.Word + " "
				classifiedText += string(word.Class) + " "
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
				prompt := widget.NewLabel(fmt.Sprintf("Classify '%s'", word))

				catBtns := make([]fyne.CanvasObject, len(parser.AvailableCategories))
				for i, cat := range parser.AvailableCategories {
					catBtns[i] = widget.NewButton(string(cat), func() {
						userChosenClassCh <- cat
						newWordsCount++
						newWordsCountLabel.SetText(fmt.Sprintf("New Words Classified: %d", newWordsCount))
						dia.Hide()
					})
				}

				diaContent := container.NewVBox(
					append([]fyne.CanvasObject{prompt}, catBtns...)...,
				)

				dia = dialog.NewCustomWithoutButtons("Classify Word", diaContent, w)
				dia.Show()
			}
		}
	}

	go updateLoop()
	w.ShowAndRun()
}

func startup(w fyne.Window, categoryCh <-chan parser.Wordcategory) *parser.Parser {
	var exPath string
	ex, err := os.Getwd()
	if err != nil {
		exPath = "./"
	} else {
		exPath = ex + "/"
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
		p, err := parser.NewParser(dictPath, inputFilePaths, outputPath, categoryCh)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			return p
		}
	}
}
