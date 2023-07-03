package main

import "os"
import "fmt"
import "bufio"
import "strconv"
import "strings"

import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

var mode int
var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}
var ROWS, COLS int
var offsetX, offsetY, currentRow, currentCol int
var source_file string
var modified bool

func read_file(filename string) {
  file, err := os.Open(filename)

  if err != nil {
    source_file = filename
    text_buffer = append(text_buffer, []rune{})
  }

  defer file.Close()
  scanner := bufio.NewScanner(file)
  lineNumber := 0

  for scanner.Scan() {
    line := scanner.Text()
    text_buffer = append(text_buffer, []rune{})

    for i := 0; i < len(line); i++ {
      text_buffer[lineNumber] = append(text_buffer[lineNumber], rune(line[i]))
    }

    lineNumber++
  }
  if lineNumber == 0 { text_buffer = append(text_buffer, []rune{}) }
}

func write_file(filename string) {
  file, err := os.Create(filename)
  if err != nil { fmt.Println(err) }
  writer := bufio.NewWriter(file)

  for row, line := range text_buffer {
      newLine := "\n"
      if row == len(text_buffer)-1 { newLine = "" }
      writeLine := string(line) + newLine
      _, err = writer.WriteString(writeLine)
      if err != nil { fmt.Println("Error:", err) }
  }

  modified = false
  writer.Flush()
}

func insert_rune(event termbox.Event) {
  insertRune := make([]rune, len(text_buffer[currentRow])+1)
  copy(insertRune[:currentCol], text_buffer[currentRow][:currentCol])
  if event.Key == termbox.KeySpace { insertRune[currentCol] = rune(' ')
  } else if event.Key == termbox.KeyTab { insertRune[currentCol] = rune(' ')
  } else { if rune(event.Ch) != '~' { insertRune[currentCol] = rune(event.Ch)
  } else { insertRune[currentCol] = rune('\t') }}
  copy(insertRune[currentCol+1:], text_buffer[currentRow][currentCol:])
  text_buffer[currentRow] = insertRune
  currentCol++
}

func delete_rune() {
  if currentCol > 0 {
    currentCol--
    deleteRune := make([]rune, len(text_buffer[currentRow])-1)
    copy(deleteRune[:currentCol], text_buffer[currentRow][:currentCol])
    copy(deleteRune[currentCol:], text_buffer[currentRow][currentCol+1:])
    text_buffer[currentRow] = deleteRune
  } else if currentRow > 0 {
    appendLine := make([]rune, len(text_buffer[currentRow]))
    copy(appendLine, text_buffer[currentRow][currentCol:])
    new_text_buffer := make([][]rune, len(text_buffer)-1)
    copy(new_text_buffer[:currentRow], text_buffer[:currentRow])
    copy(new_text_buffer[currentRow:], text_buffer[currentRow+1:])
    text_buffer = new_text_buffer
    currentRow--
    currentCol = len(text_buffer[currentRow])
    insertLine := make([]rune, len(text_buffer[currentRow]) + len(appendLine))
    copy(insertLine[:len(text_buffer[currentRow])], text_buffer[currentRow])
    copy(insertLine[len(text_buffer[currentRow]):], appendLine)
    text_buffer[currentRow] = insertLine
  }
}

func insert_line() {
  afterLine := make([]rune, len(text_buffer[currentRow][currentCol:]))
  copy(afterLine, text_buffer[currentRow][currentCol:])
  beforeLine := make([]rune, len(text_buffer[currentRow][:currentCol]))
  copy(beforeLine, text_buffer[currentRow][:currentCol])
  text_buffer[currentRow] = beforeLine
  currentRow++
  currentCol = 0
  new_text_buffer := make([][]rune, len(text_buffer)+1)
  copy(new_text_buffer[:currentRow], text_buffer[:currentRow])
  new_text_buffer[currentRow] = afterLine
  copy(new_text_buffer[currentRow+1:], text_buffer[currentRow:])
  text_buffer = new_text_buffer
}

func cut_line() {
  copy_line()
  if currentRow >= len(text_buffer) || len(text_buffer) < 2 { return }
  new_text_buffer := make([][]rune, len(text_buffer)-1)                                                   
  copy(new_text_buffer[:currentRow], text_buffer[:currentRow])                                            
  copy(new_text_buffer[currentRow:], text_buffer[currentRow+1:])                                             
  text_buffer = new_text_buffer
  if currentRow > 0 { currentRow--; currentCol = 0 }
}

func copy_line() {
  copy_line := make([]rune, len(text_buffer[currentRow]))
  copy(copy_line, text_buffer[currentRow])
  copy_buffer = copy_line
}

func paste_line() {
  if len(copy_buffer) == 0 { currentRow++; currentCol = 0 }
  new_text_buffer := make([][]rune, len(text_buffer)+1)               
  copy(new_text_buffer[:currentRow], text_buffer[:currentRow])        
  new_text_buffer[currentRow] = copy_buffer
  copy(new_text_buffer[currentRow+1:], text_buffer[currentRow:])      
  text_buffer = new_text_buffer
}

func push_text_buffer() {
  copy_undo_buffer := make([][]rune, len(text_buffer))
  copy(copy_undo_buffer, text_buffer)
  undo_buffer = copy_undo_buffer
}

func pull_text_buffer() {
  if len(undo_buffer) == 0 { return }
  text_buffer = undo_buffer
}

func scroll_text_buffer() {
  if currentRow < offsetY { offsetY = currentRow }
  if currentCol < offsetX { offsetX = currentCol }
  if currentRow >= offsetY + ROWS { offsetY = currentRow - ROWS+1 }
  if currentCol >= offsetX + COLS { offsetX = currentCol - COLS+1 }
}

func display_text_buffer() {
  var row, col int
  for row = 0; row < ROWS; row++ {
    text_bufferRow := row + offsetY
    for col = 0; col < COLS; col++ {
      text_bufferCol := col + offsetX
      if text_bufferRow >= 0 &&  text_bufferRow < len(text_buffer) &&
         text_bufferCol < len(text_buffer[text_bufferRow]) {
        if text_buffer[text_bufferRow][text_bufferCol] != rune('\t') {
          termbox.SetChar(col, row, text_buffer[text_bufferRow][text_bufferCol])
        } else { termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen) }
      } else if row+offsetY > len(text_buffer)-1 {
    termbox.SetCell(0, row, '*', termbox.ColorBlue, termbox.ColorDefault)}}
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
  var mode_status string
  var copy_status string
  var undo_status string
  if mode > 0 { mode_status = " EDIT: "
  } else { mode_status = " VIEW: " }
  filename_length := len(source_file)
  if filename_length > 8 { filename_length = 8 }
  file_status := source_file[:filename_length] + " - " + strconv.Itoa(len(text_buffer)) + " lines"
  if modified { file_status += " modified "
  } else { file_status += " saved" }
  cursor_status := " Row " + strconv.Itoa(currentRow+1) + ", Col " + strconv.Itoa(currentCol+1) + " "
  if len(copy_buffer) > 0 { copy_status = " [Copy]" }
  if len(undo_buffer) > 0 { undo_status = " [Undo]" }
  used_space := len(mode_status) + len(file_status) + len(cursor_status) + len(copy_status) + len (undo_status)
  spaces := strings.Repeat(" ", COLS - used_space)
  message := mode_status + file_status + copy_status + undo_status + spaces + cursor_status
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
  if keyEvent.Key == termbox.KeyEsc { mode = 0
  } else if keyEvent.Ch != 0 {
     if mode == 1 { insert_rune(keyEvent); modified = true
    } else {
      nineth_part := int((len(text_buffer)-1)/9)
      switch keyEvent.Ch {
        case 'q': termbox.Close(); os.Exit(0)
        case 'e': mode = 1
        case 'w': write_file(source_file)
        case 'd': cut_line()
        case 'c': copy_line()
        case 'p': paste_line()
        case 's': push_text_buffer()
        case 'l': pull_text_buffer()
        case '0': currentRow = 0; currentCol = 0
        case '1': currentRow = nineth_part; currentCol = 0
        case '2': currentRow = nineth_part*2; currentCol = 0
        case '3': currentRow = nineth_part*3; currentCol = 0
        case '4': currentRow = nineth_part*4; currentCol = 0
        case '5': currentRow = nineth_part*5; currentCol = 0
        case '6': currentRow = nineth_part*6; currentCol = 0
        case '7': currentRow = nineth_part*7; currentCol = 0
        case '8': currentRow = nineth_part*8; currentCol = 0
        case '9': currentRow = nineth_part*9; currentCol = 0
      }
    }
  } else {
    switch keyEvent.Key {
     case termbox.KeyTab:
       if mode == 1 {
         for i:= 0; i < 4; i++ { insert_rune(keyEvent); }
         modified = true
       }
     case termbox.KeySpace: if mode == 1 { insert_rune(keyEvent); modified = true }
     case termbox.KeyEnter: if mode == 1 { insert_line(); modified = true }
     case termbox.KeyBackspace: if mode == 1 {delete_rune(); modified = true }
     case termbox.KeyBackspace2: if mode == 1 { delete_rune(); modified = true }
     case termbox.KeyArrowUp: if currentRow != 0 { currentRow -- }
     case termbox.KeyArrowDown: if currentRow < len(text_buffer)-1 { currentRow++ }
     case termbox.KeyHome: currentCol = 0
     case termbox.KeyEnd: currentCol = len(text_buffer[currentRow])
     case termbox.KeyPgup: if currentRow - int(ROWS/4) > 0 { currentRow -= int(ROWS/4) }
     case termbox.KeyPgdn: if currentRow + int(ROWS/4) < len(text_buffer)-1 { currentRow += int(ROWS/4) }
     case termbox.KeyArrowLeft:
       if currentCol != 0 {
         currentCol --
       } else if currentRow > 0 {
         currentRow -= 1;
         currentCol = len(text_buffer[currentRow])
       }
     case termbox.KeyArrowRight:
       if currentCol < len(text_buffer[currentRow]) {
         currentCol++
       } else if currentRow < len(text_buffer)-1 {
         currentRow += 1
         currentCol = 0
       }
    }
    if currentCol > len(text_buffer[currentRow]) { currentCol = len(text_buffer[currentRow]) }
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
    scroll_text_buffer()
    display_text_buffer()    
    display_status_bar()
    termbox.SetCursor(currentCol - offsetX, currentRow - offsetY)
    termbox.Flush()
    process_keypress()
  }
}

func main() {
  run_editor()
}