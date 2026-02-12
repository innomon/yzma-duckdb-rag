package main

import (
	"io"

	"github.com/ledongthuc/pdf"
)

// ReadPDF extracts plain text from a PDF file at the given path.
// It uses github.com/ledongthuc/pdf for parsing.
func ReadPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return "", err
	}

	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
