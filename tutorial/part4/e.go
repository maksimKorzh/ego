package main

import "os"
import "fmt"
import "bufio"
import "strconv"
import "strings"
import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

var mode int
var ROWS, COLS int
var offset_col, offset_row int
var current_col, current_row int
var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}
var source_file string
var modified bool

func read_file(filename string) {
  file, err := os.Open(filename)
  if err != nil {
    source_file = filename
    text_buffer = append(text_buffer, []rune{}); return
  };defer file.Close()
  scanner := bufio.NewScanner(file)
  lineNumber := 0
  for scanner.Scan() {
    line := scanner.Text()
    text_buffer = append(text_buffer, []rune{})
    for i:= 0; i < len(line); i++ {
      text_buffer[lineNumber] = append(text_buffer[lineNumber], rune(line[i]))
    };lineNumber++
  };if lineNumber == 0 {
    text_buffer = append(text_buffer, []rune{})
  }
}

func display_text_buffer() {
  var row, col int
  for row = 0; row < ROWS; row++ {
    text_bufferRow := row + offset_row
    for col = 0; col < COLS; col++ {
      text_bufferCol := col + offset_col
      if text_bufferRow >= 0 && text_bufferRow < len(text_buffer) && text_bufferCol < len(text_buffer[text_bufferRow]) {
        if text_buffer[text_bufferRow][text_bufferCol] != '\t' {
          termbox.SetChar(col, row, text_buffer[text_bufferRow][text_bufferCol])
        } else { termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen) }
      } else if row+offset_row > len(text_buffer)-1 {
        termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)}}    
        termbox.SetChar(col, row, rune('\n'))
  }
}

func display_status_bar() {
  var mode_status string
  var file_status string
  var copy_status string
  var undo_status string
  var cursor_status string
  if mode > 0 { mode_status = " EDIT: "
  } else { mode_status = " VIEW: " }
  filename_length := len(source_file)
  if filename_length > 8 { filename_length = 8 }
  file_status = source_file[:filename_length] + " - " + strconv.Itoa(len(text_buffer)) + " lines"
  if modified { file_status += " modified" } else { file_status += " saved" }
  cursor_status = " Row " + strconv.Itoa(current_row+1) + ", Col " + strconv.Itoa(current_col+1) + " "
  if len(copy_buffer) > 0 { copy_status = " [Copy]" }
  if len(undo_buffer) > 0 { undo_status = " [Undo]" }
  used_space := len(mode_status) + len(file_status) + len(cursor_status) + len(copy_status) + len(undo_status)
  spaces := strings.Repeat(" ", COLS - used_space)
  message := mode_status + file_status + copy_status + undo_status + spaces + cursor_status
  print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)
}

func print_message(col, row int, fg, bg termbox.Attribute, message string) {
  for _, ch := range message {
    termbox.SetCell(col, row, ch, fg, bg)
    col += runewidth.RuneWidth(ch)
  }
}

func run_editor() {
  err := termbox.Init()
  if err != nil { fmt.Println(err); os.Exit(1) }
  if len(os.Args) > 1 {
    source_file = os.Args[1]
    read_file(source_file)
  } else {
    source_file = "out.txt"
    text_buffer = append(text_buffer, []rune{})
  }

  for {
    COLS, ROWS = termbox.Size(); ROWS--
    if COLS < 78 { COLS = 78 }
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    display_text_buffer()
    display_status_bar()
    termbox.Flush()
    event := termbox.PollEvent()
    if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
      termbox.Close()
      break
    }
  }
}

func main() {
  run_editor()
}