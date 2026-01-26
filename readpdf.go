package main

import (
	"io"

	"github.com/ledongthuc/pdf"
)

// Helper function to extract plain text from a PDF. Excerpted from
// https://github.com/ledongthuc/pdf
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
