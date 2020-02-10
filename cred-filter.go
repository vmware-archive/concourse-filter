package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	source := os.Stdin

	destination := NewLineWriter(os.Stdout)

	if len(os.Args) > 1 && os.Args[1] == "-stderr" {
		destination = NewLineWriter(os.Stderr)
	}

	defer destination.Close()

	redacted, maxSize := RedactedList()

	err := Stream(source, destination, redacted, maxSize)
	if err != nil {
		log.Fatal(err)
	}
}

type LineWriter struct {
	buffer []byte
	writer io.Writer
}

func NewLineWriter(writer io.Writer) *LineWriter {
	return &LineWriter{
		buffer: []byte{},
		writer: writer,
	}
}

func (lw *LineWriter) Write(p []byte) (int, error) {
	lw.buffer = append(lw.buffer, p...)

	index := bytes.IndexRune(lw.buffer, '\n')

	if index != -1 {
		_, err := lw.writer.Write(lw.buffer[:index+1])
		if err != nil {
			return 0, err
		}

		lw.buffer = lw.buffer[index+1:]
	}

	return len(p), nil
}

func (lw *LineWriter) Close() error {
	_, err := lw.writer.Write(lw.buffer)
	if err != nil {
		return err
	}

	return nil
}

type RedactedVariable struct {
	Name  string
	Value []byte
}

func RedactedList() ([]RedactedVariable, int) {
	whiteList := map[string]struct{}{}
	for _, value := range strings.Split(os.Getenv("CREDENTIAL_FILTER_WHITELIST"), ",") {
		whiteList[value] = struct{}{}
	}

	var redacted []RedactedVariable

	for _, variable := range os.Environ() {
		pair := strings.Split(variable, "=")

		if pair[1] == "" {
			continue
		}

		if _, ok := whiteList[pair[0]]; ok {
			continue
		}

		redacted = append(redacted, RedactedVariable{
			Name:  pair[0],
			Value: []byte(pair[1]),
		})
	}

	sort.Slice(redacted, func(i, j int) bool {
		return len(redacted[i].Value) > len(redacted[j].Value)
	})

	var maxSize int
	if len(redacted) > 0 {
		maxSize = len(redacted[0].Value)
	}

	return redacted, maxSize
}

func Stream(source io.Reader, destination io.WriteCloser, redacted []RedactedVariable, maxSize int) error {
	if maxSize == 0 {
		_, err := io.Copy(destination, source)
		if err != nil {
			return fmt.Errorf("failed to copy source to destination: %w", err)
		}

		return nil
	}

	reader := bufio.NewReader(source)

	preview, err := reader.Peek(maxSize)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to preview source: %w", err)
	}

	for {
		var match bool

		for _, r := range redacted {
			if bytes.HasPrefix(preview, r.Value) {
				_, err = reader.Discard(len(r.Value))
				if err != nil {
					return fmt.Errorf("failed to discard from source: %w", err)
				}

				fmt.Fprintf(destination, "[redacted %s]", r.Name)

				match = true

				break
			}
		}

		if !match {
			b, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					return nil
				}

				return fmt.Errorf("failed to read byte from source: %w", err)
			}

			_, err = destination.Write([]byte{b})
			if err != nil {
				return fmt.Errorf("failed to write byte to destination: %w", err)
			}
		}

		preview, err = reader.Peek(maxSize)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to preview source: %w", err)
		}
	}
}
