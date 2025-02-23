package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/flopp/go-findfont"
)

type QueueItem struct {
	itemType  string // "text", "image", "pad", "cutmark"
	textLines []string
	imagePath string
	padSize   string
}

type PTouchGUI struct {
	window    fyne.Window
	textLines [4]*widget.Entry
	fontSize  *widget.Entry
	padSize   *widget.Entry
	queue     []QueueItem
	queueBox  *fyne.Container
	preview   *canvas.Image
	tempFile  string
	fontName  string
	printBtn  *widget.Button
	saveBtn   *widget.Button
	debug     *widget.Entry
}

func getSystemFonts() []string {
	fonts := findfont.List()
	var filteredFonts []string
	for _, font := range fonts {
		if strings.HasSuffix(font, ".ttf") {
			filteredFonts = append(filteredFonts, strings.TrimSuffix(filepath.Base(font), ".ttf"))
		}
	}
	return filteredFonts
}

func newQueueItemWidget(gui *PTouchGUI, index int) *fyne.Container {
	item := gui.queue[index]

	var content string
	switch item.itemType {
	case "text":
		content = "Text: " + strings.Join(item.textLines, " | ")
	case "image":
		content = "Image: " + filepath.Base(item.imagePath)
	case "pad":
		content = "Padding: " + item.padSize + "px"
	case "cutmark":
		content = "Cutmark"
	}

	label := widget.NewLabel(content)
	label.Truncation = fyne.TextTruncateClip

	// Create movement buttons
	upBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		if index > 0 {
			gui.queue[index], gui.queue[index-1] = gui.queue[index-1], gui.queue[index]
			gui.refreshQueue()
		}
	})
	downBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		if index < len(gui.queue)-1 {
			gui.queue[index], gui.queue[index+1] = gui.queue[index+1], gui.queue[index]
			gui.refreshQueue()
		}
	})
	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		gui.queue = append(gui.queue[:index], gui.queue[index+1:]...)
		gui.refreshQueue()
	})

	// Disable up/down buttons at boundaries
	upBtn.Disable()
	downBtn.Disable()
	if index > 0 {
		upBtn.Enable()
	}
	if index < len(gui.queue)-1 {
		downBtn.Enable()
	}

	// Create a container for buttons
	buttons := container.NewHBox(upBtn, downBtn, deleteBtn)

	// Create the item container with the label expanding to fill space
	return container.NewBorder(nil, nil, nil, buttons, label)
}

func (gui *PTouchGUI) refreshQueue() {
	gui.queueBox.Objects = nil
	for i := range gui.queue {
		gui.queueBox.Add(newQueueItemWidget(gui, i))
	}
	gui.queueBox.Refresh()
}

func newPTouchGUI() *PTouchGUI {
	gui := &PTouchGUI{
		queue: make([]QueueItem, 0),
	}

	myApp := app.New()
	myApp.SetIcon(resourceIconSvg)
	gui.window = myApp.NewWindow("ptouch-print GUI")

	// Create text input fields
	gui.textLines = [4]*widget.Entry{}
	for i := 0; i < 4; i++ {
		gui.textLines[i] = widget.NewEntry()
		gui.textLines[i].SetPlaceHolder(fmt.Sprintf("Text line %d", i+1))
	}

	// Create other input fields
	gui.fontSize = widget.NewEntry()
	gui.padSize = widget.NewEntry()
	gui.padSize.SetPlaceHolder("Padding pixels")

	// Create font selection
	fontSelect := widget.NewSelect(getSystemFonts(), func(s string) {
		gui.fontName = s
	})
	fontSelect.PlaceHolder = "Select Font"

	// Create queue container
	gui.queueBox = container.NewVBox()

	// Create preview image
	gui.preview = &canvas.Image{}
	gui.preview.SetMinSize(fyne.NewSize(400, 100))
	gui.preview.FillMode = canvas.ImageFillOriginal

	// Create buttons
	addTextButton := widget.NewButton("Add Text to Queue", func() { gui.addTextToQueue() })
	addImageButton := widget.NewButton("Add Image to Queue", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, gui.window)
				return
			}
			if reader == nil {
				return
			}
			gui.addImageToQueue(reader.URI().Path())
		}, gui.window)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
		fd.Show()
	})
	addPadButton := widget.NewButton("Add Padding", func() { gui.addPadToQueue() })
	addCutmarkButton := widget.NewButton("Add Cutmark", func() { gui.addCutmarkToQueue() })
	previewButton := widget.NewButton("Generate Preview", func() { gui.generatePreview() })

	// Create Print and Save buttons (initially disabled)
	gui.printBtn = widget.NewButton("Print", func() { gui.print() })
	gui.saveBtn = widget.NewButton("Save PNG", func() { gui.savePNG() })
	gui.printBtn.Disable()
	gui.saveBtn.Disable()

	// Create a debug entry
	gui.debug = widget.NewMultiLineEntry()
	gui.debug.SetMinRowsVisible(5)
	gui.debug.Disable()
	gui.debug.Wrapping = fyne.TextWrapWord

	// Create info buttons
	versionBtn := widget.NewButton("Show Version", func() {
		cmd := exec.Command("ptouch-print", "--version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			dialog.ShowError(err, gui.window)
			return
		}
		dialog.ShowInformation("Version Information", string(output), gui.window)
	})

	infoBtn := widget.NewButton("Show Info", func() {
		cmd := exec.Command("ptouch-print", "--info")
		output, err := cmd.CombinedOutput()
		if err != nil {
			dialog.ShowError(err, gui.window)
			return
		}
		dialog.ShowInformation("Printer Information", string(output), gui.window)
	})

	supportedBtn := widget.NewButton("List Supported", func() {
		cmd := exec.Command("ptouch-print", "--list-supported")
		output, err := cmd.CombinedOutput()
		if err != nil {
			dialog.ShowError(err, gui.window)
			return
		}
		dialog.ShowInformation("Supported Printers", string(output), gui.window)
	})

	resetBtn := widget.NewButton("Reset All", func() {
		gui.resetAll()
	})

	// Create layouts
	textBox := container.NewVBox()
	for _, entry := range gui.textLines {
		textBox.Add(entry)
	}

	fontBox := container.NewGridWithColumns(2,
		widget.NewLabel("Font:"), fontSelect,
		widget.NewLabel("Font size:"), gui.fontSize)

	// Info buttons container at the bottom
	infoButtons := container.NewHBox(
		versionBtn, infoBtn, supportedBtn, resetBtn,
	)

	leftSide := container.NewVBox(
		widget.NewCard("Add Queue Items", "",
			container.NewVBox(
				textBox,
				addTextButton,
				addImageButton,
				container.NewGridWithColumns(2,
					widget.NewLabel("Padding:"),
					gui.padSize,
				),
				addPadButton,
				addCutmarkButton,
			),
		),
		widget.NewCard("Font Settings", "", fontBox),
		layout.NewSpacer(),
		gui.debug,
		infoButtons,
	)

	// Center the action buttons
	actionButtons := container.NewHBox(
		layout.NewSpacer(),
		previewButton,
		gui.printBtn,
		gui.saveBtn,
		layout.NewSpacer(),
	)

	rightSide := container.NewBorder(
		nil,
		widget.NewCard("Preview", "",
			container.NewVBox(
				gui.preview,
				actionButtons,
			),
		),
		nil, nil,
		widget.NewCard("Queue", "",
			container.NewBorder(
				widget.NewLabel("Items to print:"),
				nil, nil, nil,
				container.NewVScroll(gui.queueBox),
			),
		),
	)

	// Split view with queue taking more space
	split := container.NewHSplit(leftSide, rightSide)
	split.SetOffset(0.3)

	content := container.NewPadded(split)
	gui.window.SetContent(content)
	gui.window.Resize(fyne.NewSize(1000, 800))

	return gui
}

func (gui *PTouchGUI) resetAll() {
	// Clear text entries
	for _, entry := range gui.textLines {
		entry.SetText("")
	}

	// Clear other entries
	gui.fontSize.SetText("")
	gui.padSize.SetText("")
	gui.fontName = ""

	// Clear queue
	gui.queue = make([]QueueItem, 0)
	gui.refreshQueue()

	// Clear preview
	gui.preview.File = ""
	gui.preview.Refresh()
	if gui.tempFile != "" {
		os.Remove(gui.tempFile)
		gui.tempFile = ""
	}

	// Disable buttons
	gui.printBtn.Disable()
	gui.saveBtn.Disable()
}

func (gui *PTouchGUI) addTextToQueue() {
	var lines []string
	for _, entry := range gui.textLines {
		if text := entry.Text; text != "" {
			lines = append(lines, text)
		}
	}
	if len(lines) > 0 {
		gui.queue = append(gui.queue, QueueItem{
			itemType:  "text",
			textLines: lines,
		})
		gui.refreshQueue()
		// Clear text entries
		for _, entry := range gui.textLines {
			entry.SetText("")
		}
	}
}

func (gui *PTouchGUI) addImageToQueue(path string) {
	gui.queue = append(gui.queue, QueueItem{
		itemType:  "image",
		imagePath: path,
	})
	gui.refreshQueue()
}

func (gui *PTouchGUI) addPadToQueue() {
	if pad := gui.padSize.Text; pad != "" {
		gui.queue = append(gui.queue, QueueItem{
			itemType: "pad",
			padSize:  pad,
		})
		gui.refreshQueue()
	}
}

func (gui *PTouchGUI) addCutmarkToQueue() {
	gui.queue = append(gui.queue, QueueItem{
		itemType: "cutmark",
	})
	gui.refreshQueue()
}

func (gui *PTouchGUI) buildCommand(outputPath string) []string {
	var args []string

	// Add font settings if specified
	if gui.fontName != "" {
		args = append(args, "--font", gui.fontName)
	}
	if size := gui.fontSize.Text; size != "" {
		args = append(args, "--fontsize", size)
	}

	// Add output path
	args = append(args, "--writepng", outputPath)

	// Add all queued items
	for _, item := range gui.queue {
		switch item.itemType {
		case "text":
			args = append(args, "--text")
			args = append(args, item.textLines...)
		case "image":
			args = append(args, "--image", item.imagePath)
		case "pad":
			args = append(args, "--pad", item.padSize)
		case "cutmark":
			args = append(args, "--cutmark")
		}
	}
	gui.debug.Text = fmt.Sprintf("ptouch-print %s", strings.Join(args, " "))
	gui.debug.Refresh()
	return args
}

func (gui *PTouchGUI) generatePreview() {
	if len(gui.queue) == 0 {
		dialog.ShowInformation("Error", "Queue is empty", gui.window)
		return
	}

	tempFile, err := os.CreateTemp("", "ptouch-preview-*.png")
	if err != nil {
		dialog.ShowError(fmt.Errorf("error creating temp file: %v", err), gui.window)
		return
	}
	tempFile.Close()

	if gui.tempFile != "" {
		os.Remove(gui.tempFile)
	}
	gui.tempFile = tempFile.Name()

	args := gui.buildCommand(gui.tempFile)
	cmd := exec.Command("ptouch-print", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		dialog.ShowError(fmt.Errorf("preview error: %v\n%s", err, output), gui.window)
		return
	}

	gui.preview.File = gui.tempFile
	gui.preview.Refresh()

	// Enable print and save buttons
	gui.printBtn.Enable()
	gui.saveBtn.Enable()
}

func (gui *PTouchGUI) print() {
	if gui.tempFile == "" {
		dialog.ShowInformation("Error", "Please generate a preview first", gui.window)
		return
	}

	cmd := exec.Command("ptouch-print", "--image", gui.tempFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		dialog.ShowError(fmt.Errorf("printing error: %v\n%s", err, output), gui.window)
		return
	}

	dialog.ShowInformation("Success", "Print job completed successfully", gui.window)
}

func (gui *PTouchGUI) savePNG() {
	if gui.tempFile == "" {
		dialog.ShowInformation("Error", "Please generate a preview first", gui.window)
		return
	}

	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, gui.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Copy temp file to selected location
		source, err := os.Open(gui.tempFile)
		if err != nil {
			dialog.ShowError(err, gui.window)
			return
		}
		defer source.Close()

		_, err = io.Copy(writer, source)
		if err != nil {
			dialog.ShowError(err, gui.window)
			return
		}

		dialog.ShowInformation("Success", "File saved successfully", gui.window)
	}, gui.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
	fd.Show()
}

func main() {
	gui := newPTouchGUI()
	gui.window.Resize(fyne.NewSize(1000, 800))
	gui.window.ShowAndRun()
}
