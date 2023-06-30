package main

import "os"
import "bufio"
import "fmt"
import "strconv"
import "strings"

import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

var buffer = [][]rune{}
var ROWS, COLS int
var offsetX, offsetY, currentRow, currentCol int
var source_file string
var modified bool

func read_file(filename string) {
	file, err := os.Open(filename)

	if err != nil {
	  source_file = filename
	  buffer = append(buffer, []rune{})
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
  lineNumber := 0

	for scanner.Scan() {
		line := scanner.Text()
		buffer = append(buffer, []rune{})

		for i := 0; i < len(line); i++ {
		  buffer[lineNumber] = append(buffer[lineNumber], rune(line[i]))
		}

		lineNumber++
	}
}

func write_file(filename string) {
  file, err := os.Create(filename)
  if err != nil { fmt.Println(err) }
  writer := bufio.NewWriter(file)

  for row, line := range buffer {
      newLine := "\n"
      if row == len(buffer)-1 { newLine = "" }
      writeLine := string(line) + newLine
      _, err = writer.WriteString(writeLine)
      if err != nil { fmt.Println("Error:", err) }
  }
  modified = false
  writer.Flush()
}

func insert_rune(event termbox.Event) {
  insertRune := make([]rune, len(buffer[currentRow])+1)
  copy(insertRune[:currentCol], buffer[currentRow][:currentCol])
  if event.Key == termbox.KeySpace { insertRune[currentCol] = rune(' ')
  } else { insertRune[currentCol] = rune(event.Ch) }
  copy(insertRune[currentCol+1:], buffer[currentRow][currentCol:])
  buffer[currentRow] = insertRune
  currentCol++
}

func delete_rune() {
  if currentCol > 0 {
    currentCol--
    deleteRune := make([]rune, len(buffer[currentRow])-1)
    copy(deleteRune[:currentCol], buffer[currentRow][:currentCol])
    copy(deleteRune[currentCol:], buffer[currentRow][currentCol+1:])
    buffer[currentRow] = deleteRune
  } else if currentRow > 0 {
    appendLine := make([]rune, len(buffer[currentRow]))
    copy(appendLine, buffer[currentRow][currentCol:])
    newBuffer := make([][]rune, len(buffer)-1)
    copy(newBuffer[:currentRow], buffer[:currentRow])
    copy(newBuffer[currentRow:], buffer[currentRow+1:])
    buffer = newBuffer
    currentRow--
    currentCol = len(buffer[currentRow])
    insertLine := make([]rune, len(buffer[currentRow]) + len(appendLine))
    copy(insertLine[:len(buffer[currentRow])], buffer[currentRow])
    copy(insertLine[len(buffer[currentRow]):], appendLine)
    buffer[currentRow] = insertLine
  }
}

func insert_line() {
  afterLine := make([]rune, len(buffer[currentRow][currentCol:]))
  copy(afterLine, buffer[currentRow][currentCol:])
  beforeLine := make([]rune, len(buffer[currentRow][:currentCol]))
  copy(beforeLine, buffer[currentRow][:currentCol])
  buffer[currentRow] = beforeLine
  currentRow++
  currentCol = 0
  newBuffer := make([][]rune, len(buffer)+1)
  copy(newBuffer[:currentRow], buffer[:currentRow])
  newBuffer[currentRow] = afterLine
  copy(newBuffer[currentRow+1:], buffer[currentRow:])
  buffer = newBuffer
}

func scroll_buffer() {
  if currentRow < offsetY { offsetY = currentRow }
  if currentCol < offsetX { offsetX = currentCol }
  if currentRow >= offsetY + ROWS { offsetY = currentRow - ROWS+1 }
  if currentCol >= offsetX + COLS { offsetX = currentCol - COLS+1 }
}

func display_buffer() {
  var row, col int
  for row = 0; row < ROWS; row++ {
    bufferRow := row + offsetY
    for col = 0; col < COLS; col++ {
      bufferCol := col + offsetX
      if bufferRow >= 0 &&  bufferRow < len(buffer) &&
         bufferCol < len(buffer[bufferRow]) {
        termbox.SetChar(col, row, buffer[bufferRow][bufferCol])
      }}
    termbox.SetChar(col, row, '\n')
  }
}

func print_message(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func display_status_bar() {
  file_status := " " + source_file + " - " + strconv.Itoa(len(buffer)) + " lines"
  if modified { file_status += " modified "
  } else { file_status += " saved" }
  cursor := " Row " + strconv.Itoa(currentRow) + ", Col " + strconv.Itoa(currentCol) + " "
  spaces := strings.Repeat(" ", COLS - len(file_status) - len(cursor))
  message := file_status + spaces + cursor
  print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)
}

func get_key() termbox.Event {
  var keyEvent termbox.Event
  switch event := termbox.PollEvent(); event.Type {
	   case termbox.EventKey: keyEvent = event
	   case termbox.EventError: panic(event.Err)
 	}
	 return keyEvent
}

func process_keypress() {
  keyEvent := get_key()
  if keyEvent.Key == termbox.KeyEsc {
    termbox.Close()
    os.Exit(0)
  } else if keyEvent.Ch != 0 {
	  insert_rune(keyEvent); modified = true
  } else {
	  switch keyEvent.Key {
     case termbox.KeySpace: insert_rune(keyEvent); modified = true
     case termbox.KeyEnter: insert_line(); modified = true
     case termbox.KeyBackspace: delete_rune(); modified = true
     case termbox.KeyBackspace2: delete_rune(); modified = true
	    case termbox.KeyArrowUp: if currentRow != 0 { currentRow -- }
	    case termbox.KeyArrowDown: if currentRow < len(buffer)-1 { currentRow++ }
	    case termbox.KeyHome: currentCol = 0
	    case termbox.KeyEnd: currentCol = len(buffer[currentRow])
	    case termbox.KeyPgup: if currentRow - int(ROWS/2) > 0 { currentRow -= int(ROWS/2) }
	    case termbox.KeyPgdn: if currentRow + int(ROWS/2) < len(buffer)-1 { currentRow += int(ROWS/2) }
	    case termbox.KeyArrowLeft:
	      if currentCol != 0 {
	        currentCol --
	      } else if currentRow > 0 {
	        currentRow -= 1;
	        currentCol = len(buffer[currentRow])
	      }
	    case termbox.KeyArrowRight:
	      if currentCol < len(buffer[currentRow]) {
	        currentCol++
	      } else if currentRow < len(buffer)-1 {
	        currentRow += 1
	        currentCol = 0
	      }
	    case termbox.KeyCtrlS: write_file(source_file)
	  }
	  if currentCol > len(buffer[currentRow]) { currentCol = len(buffer[currentRow]) }
  }
}

// MAIN
func run_editor() {
  err := termbox.Init()
	if err != nil { fmt.Println(err); os.Exit(1) }
  if len(os.Args) > 1 {
    source_file = os.Args[1]
    read_file(source_file)
  } else {
    source_file = "noname.txt"
    buffer = append(buffer, []rune{})
  }

  for {    
    COLS, ROWS = termbox.Size(); ROWS--;
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    scroll_buffer()
    display_buffer()    
    display_status_bar()
	  termbox.SetCursor(currentCol - offsetX, currentRow - offsetY)
	  termbox.Flush()
	  process_keypress()
  }
}

func main() {
	run_editor()
}