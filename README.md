# EGO
A bare bones text editor written in Go

# Project goals
1. Create a text editor for myself
2. Learn Go programming language
3. Share knowledge with others

# Features
 - modes (VIEW/EDIT)
 - display buffer & status bar
 - inserting characters
 - deleting characters
 - inserting lines
 - deleting lines
 - navigation
 - copy/paste
 - undo/redo

# Key bindigns
       ESC: enter the 'VIEW' mode
         e: enter the 'EDIT' mode
         q: quit from the text editor
         w: write file to disk
         d: cut current line to copy buffer
         c: copy current line to copy buffer
         p: paste line from copy buffer
         s: push text buffer to undo buffer
         l: pull text buffer from undo buffer
       0-9: navigate to the begining of the file
    Arrows: move cursor
    PgDown: scroll 1/4 of the screen downwards
      PgUp: scroll 1/4 of the screen upwards
      HOME: move cursor to the begining of the current line
       END: move cursor to the end of the current line

# Video Tutorial Series
[![IMAGE ALT TEXT HERE](https://img.youtube.com/vi/mVFXBZUBe2s/0.jpg)](https://www.youtube.com/watch?v=mVFXBZUBe2s&list=PLLfIBXQeu3aa0NI4RT5OuRQsLo6gtLwGN)

# Release
https://github.com/maksimKorzh/ego/releases/tag/0.1

# Build from sources
```bash
go mod init ego
go build -o ego e.go
```