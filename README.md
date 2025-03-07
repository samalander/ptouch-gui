# ptouch-gui
A GUI for the ptouch-print command-line application. Written in Go using the Fyne toolkit.

The ptouch-print command-line application can be found here:

  - https://github.com/HenrikBengtsson/brother-ptouch-label-printer-on-linux
  - https://git.familie-radermacher.ch/linux/ptouch-print.git

It needs to be in the path so that the GUI can access it. The printer also needs to be online and accessible by the user running the GUI.

## Usage

Add items to the queue to get them printed. The various items you can create are:

  - 1-4 lines of text
  - a PNG image
  - some padding
  - a cutmark

Once the queue is filled, generate a preview and print it or save it for later once you are satisfied.

### Text

The text will be generated using the specified font (if any) at the given font size (if any). If no font is chosen, the default will be used. If no font size is specified, the text will be auto-sized to fill the tape's height. All text in the queue will be printed using the given font and font size. The 1-4 lines will be printed one under the other.

### Image

When adding a PNG image, a file browser dialog will be displayed.

This function can be used to print images previously generated by this program. For example, if you wish to print text using different fonts, you can output them to images first and then bring the images together in a new queue to print them out.

### Padding

This will insert the specified number of pixels as padding.

### Cutmark

This will insert a cutmark (vertical line).

## Queue

Items in the queue can be deleted by clicking their trash button. They can also be brought forward or pushed back in the order by clicking their up and down arrow buttons.

## Information

Everytime the queue is processed, the command-line that was executed can be viewed at the bottom of the left toolbar.

Other printer information can be viewed by clicking the buttons, also located at the bottom of the left toolbar. The program state can also be cleared using the relevant button there.