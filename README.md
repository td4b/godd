# godd
A Simple block copy and write utility for low level systems programming.

# Installation

```bash
go install github.com/td4b/godd@latest
```

Update your path or copy the binary to where you want to use it.

```bash
export PATH=$PATH:$HOME/go/bin/godd
```

# Example usage

```bash
Usage:
  go run dd.go -if=input.txt -of=output.txt -bs=1024 -workers=4

Flags:
  -bs string
        Block size in bytes (default: 512)
  -if string
        Input file (required)
  -of string
        Output file (required)
  -workers int
        Number of concurrent workers (default: 4) (default 4)
```

Example of copying a source file descriptor to a destination.
```bash
godd -if=source.txt -of=dest.txt
```
