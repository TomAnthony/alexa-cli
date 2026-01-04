package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Formatter handles output formatting for CLI commands
type Formatter struct {
	w      io.Writer
	asJSON bool
}

// NewFormatter creates a new output formatter
func NewFormatter(w io.Writer, asJSON bool) *Formatter {
	return &Formatter{w: w, asJSON: asJSON}
}

// Success writes a success response
func (f *Formatter) Success(message string) error {
	if f.asJSON {
		return f.writeJSON(map[string]interface{}{
			"success": true,
			"message": message,
		})
	}
	_, err := fmt.Fprintln(f.w, message)
	return err
}

// Error writes an error response
func (f *Formatter) Error(err error) error {
	if f.asJSON {
		return f.writeJSON(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	_, werr := fmt.Fprintf(f.w, "Error: %s\n", err.Error())
	return werr
}

// Data writes arbitrary data
func (f *Formatter) Data(data interface{}) error {
	if f.asJSON {
		return f.writeJSON(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	}
	_, err := fmt.Fprintf(f.w, "%+v\n", data)
	return err
}

// Devices writes a list of devices
func (f *Formatter) Devices(devices []Device) error {
	if f.asJSON {
		return f.writeJSON(map[string]interface{}{
			"success": true,
			"data":    devices,
		})
	}

	if len(devices) == 0 {
		_, err := fmt.Fprintln(f.w, "No devices found")
		return err
	}

	for _, d := range devices {
		online := ""
		if !d.Online {
			online = " (offline)"
		}
		_, err := fmt.Fprintf(f.w, "%-30s %s%s\n", d.Name, d.Serial, online)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) writeJSON(v interface{}) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Device represents an Alexa device for output
type Device struct {
	Name   string `json:"name"`
	Serial string `json:"serial"`
	Type   string `json:"type"`
	Family string `json:"family"`
	Online bool   `json:"online"`
}
