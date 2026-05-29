package ghactions

import (
	"errors"
	"os"
	"strings"
)

// RenderOutputs formats ordered pairs as `key=value` lines with a trailing newline — the format
// GitHub Actions reads from the $GITHUB_OUTPUT file.
func RenderOutputs(outputs []Pair) string {
	rows := make([]string, len(outputs))
	for i, p := range outputs {
		rows[i] = p.Key + "=" + p.Value
	}
	return strings.Join(rows, "\n") + "\n"
}

// AppendOutput appends rendered outputs to the $GITHUB_OUTPUT file. The path is a GHA-provided
// sink (global state), not a parameter; it errors if unset so a misconfigured action fails loudly.
func AppendOutput(path string, outputs []Pair) error {
	if path == "" {
		return errors.New("GITHUB_OUTPUT is not set")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(RenderOutputs(outputs))
	return err
}
