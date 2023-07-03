package main

import "os"
import "fmt"

import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

func print_message(col, row int, fg, bg termbox.Attribute, message string) {
  for _, ch := range message {
    termbox.SetCell(col, row, ch, fg, bg)
    col += runewidth.RuneWidth(ch)
  }
}

func run_editor() {
  err := termbox.Init()
  if err != nil { fmt.Println(err); os.Exit(1) }
  for {
    print_message(25, 11, termbox.ColorDefault, termbox.ColorDefault, "EGO - A bare bones text editor")
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