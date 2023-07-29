package main

import "os"
import "fmt"
import "bufio"
import "bytes"
import "strconv"
import "strings"
import "unicode"
import "os/exec"
import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

var mode = 0
var line_number_width = 0
var highlight = 1
var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}
var ROWS, COLS int
var offset_col, offset_row int
var current_row, current_col int
var source_file string
var modified bool

func read_file(filename string) {
  file, err := os.Open(filename)
  if err != nil {
    source_file = filename
    text_buffer = append(text_buffer, []rune{}); return
  };defer file.Close()
  scanner := bufio.NewScanner(file)
  line_number := 0
  for scanner.Scan() {
    line := scanner.Text()
    text_buffer = append(text_buffer, []rune{})
    count := 0; for _, ch := range line {
      text_buffer[line_number] = append(text_buffer[line_number], rune(ch))
      count += runewidth.RuneWidth(ch)
    };line_number++
  };if line_number == 0 { text_buffer = append(text_buffer, []rune{}) }
}

func read_stream(buffer string) {
  text_buffer = [][]rune{}
  line_number := 0
  for _, line := range strings.Split(buffer, "\n") {
    modified = true
    text_buffer = append(text_buffer, []rune{})
    count := 0; for _, ch := range line {
      if ch == '\r' { continue }
      text_buffer[line_number] = append(text_buffer[line_number], rune(ch))
      count += runewidth.RuneWidth(ch)
    };line_number++
  }
}

func write_file(filename string) {
  file, err := os.Create(filename)
  if err != nil { fmt.Println(err) }
  defer file.Close()
  writer := bufio.NewWriter(file)
  for row, line := range text_buffer {
      new_line := "\n"
      if row == len(text_buffer) { new_line = "" }
      write_line := string(line) + new_line
      _, err = writer.WriteString(write_line)
      if err != nil { fmt.Println("Error:", err) }
  }; writer.Flush(); modified = false;
}

func insert_rune(event termbox.Event) {
  insert_line := make([]rune, len(text_buffer[current_row])+1)
  copy(insert_line[:current_col], text_buffer[current_row][:current_col])
  if event.Key == termbox.KeySpace { insert_line[current_col] = rune(' ')
  } else if event.Key == termbox.KeyTab { insert_line[current_col] = rune(' ')
  } else { insert_line[current_col] = rune(event.Ch) }
  copy(insert_line[current_col+1:], text_buffer[current_row][current_col:])
  text_buffer[current_row] = insert_line
  current_col++
}

func delete_rune() {
  if current_col > 0 { current_col--
    delete_line := make([]rune, len(text_buffer[current_row])-1)
    copy(delete_line[:current_col], text_buffer[current_row][:current_col])
    copy(delete_line[current_col:], text_buffer[current_row][current_col+1:])
    text_buffer[current_row] = delete_line
  } else if current_row > 0 {
    append_line := make([]rune, len(text_buffer[current_row]))
    copy(append_line, text_buffer[current_row][current_col:])
    new_text_buffer := make([][]rune, len(text_buffer)-1)
    copy(new_text_buffer[:current_row], text_buffer[:current_row])
    copy(new_text_buffer[current_row:], text_buffer[current_row+1:])
    text_buffer = new_text_buffer;current_row--
    current_col = len(text_buffer[current_row])
    insert_line := make([]rune, len(text_buffer[current_row]) + len(append_line))
    copy(insert_line[:len(text_buffer[current_row])], text_buffer[current_row])
    copy(insert_line[len(text_buffer[current_row]):], append_line)
    text_buffer[current_row] = insert_line
  }
}

func insert_line() {
  right_line := make([]rune, len(text_buffer[current_row][current_col:]))
  copy(right_line, text_buffer[current_row][current_col:])
  left_line := make([]rune, len(text_buffer[current_row][:current_col]))
  copy(left_line, text_buffer[current_row][:current_col])
  text_buffer[current_row] = left_line
  current_row++; current_col = 0
  new_text_buffer := make([][]rune, len(text_buffer)+1)
  copy(new_text_buffer[:current_row], text_buffer[:current_row])
  new_text_buffer[current_row] = right_line
  copy(new_text_buffer[current_row+1:], text_buffer[current_row:])
  text_buffer = new_text_buffer
}

func cat_line() {
  for i := current_col; i > 0; i-- { delete_rune() }
  if current_row > 0 { delete_rune() }
}

func cut_line() {
  copy_line()
  if current_row >= len(text_buffer) || len(text_buffer) < 2 { return }
  new_text_buffer := make([][]rune, len(text_buffer)-1)
  copy(new_text_buffer[:current_row], text_buffer[:current_row])
  copy(new_text_buffer[current_row:], text_buffer[current_row+1:])
  text_buffer = new_text_buffer
  if current_row > 0 { current_row--; current_col = 0 }
}

func copy_line() {
  copy_line := make([]rune, len(text_buffer[current_row]))
  copy(copy_line, text_buffer[current_row])
  copy_buffer = copy_line
}

func paste_line() {
  if len(copy_buffer) == 0 { current_row++; current_col = 0 }
  new_text_buffer := make([][]rune, len(text_buffer)+1)
  copy(new_text_buffer[:current_row], text_buffer[:current_row])
  new_text_buffer[current_row] = copy_buffer
  copy(new_text_buffer[current_row+1:], text_buffer[current_row:])
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
  if current_row < offset_row { offset_row = current_row }
  if current_col < offset_col { offset_col = current_col }
  if current_row >= offset_row + ROWS { offset_row = current_row - ROWS+1 }
  if current_col >= offset_col + COLS-line_number_width { offset_col = current_col - COLS+line_number_width+1 }
}

func highlight_keyword(keyword string, col, row int) {
  for i := 0; i < len(keyword); i++ {
    ch := text_buffer[row+offset_row][col+offset_col+i]
    termbox.SetCell(col+i+line_number_width, row, ch, termbox.ColorGreen | termbox.AttrBold, termbox.ColorDefault);
  }
}

func highlight_string(col, row int) int {
  i := 0; for {
    if col+offset_col+i == len(text_buffer[row+offset_row]) { return i-1 }
    ch := text_buffer[row+offset_row][col+offset_col+i]
    if ch == '"' || ch == '\'' { termbox.SetCell(col+i+line_number_width, row, ch, termbox.ColorYellow, termbox.ColorDefault); return i
    } else { termbox.SetCell(col+i+line_number_width, row, ch, termbox.ColorYellow, termbox.ColorDefault); i++ }
  }
}

func highlight_comment(col, row int) int {
  i := 0; for {
    if col+offset_col+i == len(text_buffer[row+offset_row]) { return i-1 }
    ch := text_buffer[row+offset_row][col+offset_col+i]
    termbox.SetCell(col+i+line_number_width, row, ch, termbox.ColorMagenta | termbox.AttrBold, termbox.ColorDefault); i++
  }
}

func highlight_syntax(col *int, row, text_buffer_col, text_buffer_row int) {
  ch := text_buffer[text_buffer_row][text_buffer_col]
  next_token := string(text_buffer[text_buffer_row][text_buffer_col:])
  if unicode.IsDigit(ch) {
    termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorYellow | termbox.AttrBold, termbox.ColorDefault)
  } else if ch == '\'' {
    termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorYellow, termbox.ColorDefault)
    *col++; *col += highlight_string(*col, row);
  } else if ch == '"' {
    termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorYellow, termbox.ColorDefault)
    *col++; *col += highlight_string(*col, row);
  } else if strings.Contains("+-*><=%&|^!:", string(ch)) {
    termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorMagenta | termbox.AttrBold, termbox.ColorDefault)
  } else if ch == '/' {
    termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorMagenta | termbox.AttrBold, termbox.ColorDefault)
    index := strings.Index(next_token, "//")
    if index == 0 { *col += highlight_comment(*col, row) }
    index = strings.Index(next_token, "/*")
    if index == 0 { *col += highlight_comment(*col, row) }
  } else if ch == '#' {
    termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorMagenta | termbox.AttrBold, termbox.ColorDefault)
    *col += highlight_comment(*col, row)
  } else {
    for _,token := range []string{
      "false", "False", "NaN", "None", "bool", "break", "byte",
      "case", "catch", "class", "const", "continue", "def", "do",
      "elif", "else", "else:", "enum", "export", "extends", "extern",
      "finally", "float", "for", "from", "func", "function",
      "global", "if", "import", " in", "int", "lambda", "try:", "except:",
      "nil", "not", "null", "pass", "print", "raise", "return",
      "self", "short", "signed", "sizeof", "static", "struct", "switch",
      "this", "throw", "throws", "true", "True", "typedef", "typeof",
      "undefined", "union", "unsigned", "until", "var", "void",
      "while", "with", "yield", "double",
    } { index := strings.Index(next_token, token + " ")
     if index == 0 { highlight_keyword(token[:len(token)], *col, row); *col += len(token); break
      } else { termbox.SetCell(*col+line_number_width, row, ch, termbox.ColorDefault, termbox.ColorDefault) }
    }
  }
}

func display_text_buffer() {
  var row, col int
  for row = 0; row < ROWS; row++ {
    text_buffer_row := row + offset_row
    for col = 0; col < COLS; col++ {
      text_buffer_col := col + offset_col
      if text_buffer_row < len(text_buffer) {
        line_number_offset := line_number_width - len(strconv.Itoa(text_buffer_row+1))-1
        print_message(line_number_offset, row,
        termbox.ColorCyan, termbox.ColorDefault, strconv.Itoa(text_buffer_row+1))
      };if text_buffer_row >= 0 &&  text_buffer_row < len(text_buffer) &&
         text_buffer_col < len(text_buffer[text_buffer_row]) {
        if text_buffer[text_buffer_row][text_buffer_col] != rune('\t') {
          if highlight == 1 { highlight_syntax(&col, row, text_buffer_col, text_buffer_row)
          } else { termbox.SetCell(col+line_number_width, row, text_buffer[text_buffer_row][text_buffer_col],
                   termbox.ColorDefault, termbox.ColorDefault) }
        } else { termbox.SetCell(col+line_number_width, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen) }
      } else if row+offset_row > len(text_buffer)-1 {
    termbox.SetCell(0, row, '*', termbox.ColorBlue, termbox.ColorDefault)}}
    if row == current_row - offset_row && highlight == 1 {
      COLS, ROWS := termbox.Size(); ROWS--
      if row >= ROWS { continue }
      for col = 0; col < COLS; col++ {
        current_cell := termbox.GetCell(col, row)
        termbox.SetCell(col, row, current_cell.Ch, termbox.ColorDefault, termbox.ColorBlue)
      }
    }
    termbox.SetChar(col, row, '\n')
  }
}

func display_status_bar() {
  var mode_status string
  var copy_status string
  var undo_status string
  if mode > 0 { mode_status = " EDIT: "
  } else { mode_status = " VIEW: " }
  filename_length := len(source_file)
  if filename_length > 24 { filename_length = 24 }
  file_status := source_file[:filename_length] + " - " + strconv.Itoa(len(text_buffer)) + " lines"
  if modified { file_status += " modified "
  } else { file_status += " saved" }
  cursor_status := " Row " + strconv.Itoa(current_row+1) + ", Col " + strconv.Itoa(current_col+1) + " "
  if len(copy_buffer) > 0 { copy_status = " [Copy]" }
  if len(undo_buffer) > 0 { undo_status = " [Undo]" }
  used_space := len(mode_status) + len(file_status) + len(cursor_status) + len(copy_status) + len (undo_status)
  spaces := strings.Repeat(" ", COLS - used_space)
  message := mode_status + file_status + copy_status + undo_status + spaces + cursor_status
  print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)
}

func print_message(x, y int, fg, bg termbox.Attribute, msg string) {
  for _, c := range msg {
    termbox.SetCell(x, y, c, fg, bg)
    x += runewidth.RuneWidth(c)
  }
}

func get_key() termbox.Event {
  var key_event termbox.Event
  switch event := termbox.PollEvent(); event.Type {
     case termbox.EventKey: key_event = event
     case termbox.EventError: panic(event.Err)
   };return key_event
}

func execute_command() { ROWS--
  termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
  display_text_buffer()
  display_status_bar()
  termbox.SetCursor(0, ROWS+1)
  termbox.Flush()
  command := ""
  command_loop:
  for {
    event := get_key()
    switch event.Key {
      case termbox.KeyEsc: break command_loop
      case termbox.KeyEnter:
        content := ""
        for _, line := range text_buffer { content += string(line) + "\n" }
        is_search := false
        if strings.ContainsRune(command, '=') { is_search = true }
        cmd := exec.Command("sed", command)
        if is_search == true { cmd = exec.Command("sed", "-n", command) }
        cmd.Stdin = bytes.NewBufferString(content)
        var output bytes.Buffer
        cmd.Stdout = &output
        err := cmd.Run()
        if err != nil { continue }
        result := output.String()
        if len(result) > 0 {
          if is_search == true {
            current_row, err = strconv.Atoi(strings.TrimSpace(strings.Split(result, "\n")[0]))
            current_row--; current_col = 0
            break command_loop
          }
          read_stream(result[:len(result)-1])
        };if current_row > len(text_buffer)-1 { current_row = len(text_buffer)-1 }
        current_col = 0
        break command_loop
      case termbox.KeySpace: command += " "
      case termbox.KeyBackspace: if len(command) > 0 { command = command[:len(command)-1] }
      case termbox.KeyBackspace2: if len(command) > 0 { command = command[:len(command)-1] }
    };if event.Ch != 0 {
      command += string(event.Ch)
      print_message(0, ROWS+1, termbox.ColorDefault, termbox.ColorDefault, command)
    };
    command_length := 0
    for _,ch := range command { if ch > 0 { command_length++} }
    termbox.SetCursor(command_length, ROWS+1)
    for i := len(command); i < COLS; i++ { termbox.SetChar(i, ROWS+1, rune(' ')) }
    termbox.Flush()
  };ROWS++
}

func process_keypress() {
  key_event := get_key()
  if key_event.Key == termbox.KeyEsc { mode = 0
  } else if key_event.Ch != 0 {
    if mode == 1 { insert_rune(key_event); modified = true
    } else {
      eight_len := int((len(text_buffer)-1)/8)
      switch key_event.Ch {
        case 'q': termbox.Close(); os.Exit(0)
        case 'e': mode = 1
        case 'x': execute_command()
        case 'w': write_file(source_file)
        case 'a': cat_line()
        case 'd': cut_line()
        case 'c': copy_line()
        case 'p': paste_line()
        case 's': push_text_buffer()
        case 'l': pull_text_buffer()
        case 'h': highlight ^= 1
        case '0': current_row = 0; current_col = 0
        case '1': current_row = eight_len; current_col = 0
        case '2': current_row = eight_len*2; current_col = 0
        case '3': current_row = eight_len*3; current_col = 0
        case '4': current_row = eight_len*4; current_col = 0
        case '5': current_row = eight_len*5; current_col = 0
        case '6': current_row = eight_len*6; current_col = 0
        case '7': current_row = eight_len*7; current_col = 0
        case '8': current_row = eight_len*8; current_col = 0
        case '9': current_row = len(text_buffer)-1; current_col = 0
      }
    }
  } else {
    switch key_event.Key {
     case termbox.KeyTab:
       if mode == 1 {
         for i:= 0; i < 4; i++ { insert_rune(key_event); }
         modified = true
       }
     case termbox.KeySpace: if mode == 1 { insert_rune(key_event); modified = true }
     case termbox.KeyEnter: if mode == 1 { insert_line(); modified = true }
     case termbox.KeyBackspace: if mode == 1 {delete_rune(); modified = true }
     case termbox.KeyBackspace2: if mode == 1 { delete_rune(); modified = true }
     case termbox.KeyArrowUp: if current_row != 0 { current_row -- }
     case termbox.KeyArrowDown: if current_row < len(text_buffer)-1 { current_row++ }
     case termbox.KeyHome: current_col = 0
     case termbox.KeyEnd: current_col = len(text_buffer[current_row])
     case termbox.KeyPgup: if current_row - int(ROWS/4) > 0 { current_row -= int(ROWS/4) }
     case termbox.KeyPgdn: if current_row + int(ROWS/4) < len(text_buffer)-1 { current_row += int(ROWS/4) }
     case termbox.KeyArrowLeft:
       if current_col != 0 {
         current_col --
       } else if current_row > 0 {
         current_row -= 1;
         current_col = len(text_buffer[current_row])
       }
     case termbox.KeyArrowRight:
       if current_col < len(text_buffer[current_row]) {
         current_col++
       } else if current_row < len(text_buffer)-1 {
         current_row += 1
         current_col = 0
       }
    };if current_col > len(text_buffer[current_row]) { current_col = len(text_buffer[current_row]) }
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
  };for {
    line_number_width = len(strconv.Itoa(len(text_buffer)))+1
    COLS, ROWS = termbox.Size(); ROWS--;
    if COLS < 78 { COLS = 78 }
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    scroll_text_buffer()
    display_text_buffer()
    display_status_bar()
    termbox.SetCursor(current_col - offset_col+line_number_width, current_row - offset_row)
    termbox.Flush()
    process_keypress()
  }
}

func main() {
  run_editor()
}
