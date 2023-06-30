package main

import "os"
import "bufio"
//import "time"
import "fmt"
//import "strconv"

import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

//var buffer = [][]rune{{'a', 'b', 'c'}, {'1', '2', '3'}}
var buffer = [][]rune{}
var ROWS, COLS int
var offsetX, offsetY, currentRow, currentCol int
var src string

func read_file(filename string) {
	file, err := os.Open(filename)
	if err != nil { fmt.Println("Error:", err); return }
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

	if err := scanner.Err(); err != nil { fmt.Println("Error:", err) }
}

func scroll_buffer() {
  if currentRow < offsetY { offsetY = currentRow }
  if currentCol < offsetX { offsetX = currentCol }
  if currentRow >= offsetY + ROWS { offsetY = currentRow - ROWS }
  if currentCol >= offsetX + COLS { offsetX = currentCol - COLS }
}

func display_buffer() {
  ROWS, COLS = termbox.Size()
  var row, col int
  for row = 0; row < ROWS; row++ {
    bufferRow := row + offsetY
    for col = 0; col < COLS; col++ {
      bufferCol := col + offsetX
      if row < len(buffer) && col < len(buffer[row]) {
        termbox.SetChar(col, row, buffer[bufferRow][bufferCol])
      }
    }
    termbox.SetChar(col, row, '\n')
  }
  termbox.SetCursor(currentCol - offsetX, currentRow - offsetY)
  termbox.Flush()
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
  } else if keyEvent.Ch != 0 { // printable characters
	  fmt.Printf("Typed: %c\n", rune(keyEvent.Ch))
  } else { // non-printable characters
	  switch keyEvent.Key {
	    case termbox.KeyArrowUp: if currentRow != 0 { currentRow -- }
	    case termbox.KeyArrowDown: if currentRow < len(buffer)-1 { currentRow++ }
	  }
  }
}

func run_editor() {
  for {    
    scroll_buffer()
    display_buffer()
	  process_keypress()
  }
}


func main() {
	err := termbox.Init()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

  read_file("e.go")
  run_editor()
	
	
}



// This function is often useful:
func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}






