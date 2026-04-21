package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

const (
	BROTHER_QL_MODEL = "QL-700"

	STANDARD_PRINTER    = "usb://0x04f9:2042?serial=000F2G709185"
	STANDARD_LABEL_SIZE = "29x90"

	SMALL_PRINTER    = "usb://0x04f9:2042?serial=SMALL_SERIAL"
	SMALL_LABEL_SIZE = "17x54"
)

func printHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if the json is malformed
	var request Request
	if malformedJson(&request, w, r) {
		return
	}

	// Valdate the json
	if !validJson(&request, w) {
		return
	}

	var labelFile string
	var labelSize string
	var printer string

	switch request.LabelType {
	case 1: // small
		labelFile = "temp/label_small.png"
		labelSize = SMALL_LABEL_SIZE
		printer = SMALL_PRINTER
		if err := formatSmallLabel(request.ItemId, request.Serial, request.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Ok: false, Error: err.Error()})
			return
		}
	case 2: // small cable
		labelFile = "temp/label_small_cable.png"
		labelSize = SMALL_LABEL_SIZE
		printer = SMALL_PRINTER
		if err := formatSmallCableLabel(request.ItemId, request.Serial, request.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Ok: false, Error: err.Error()})
			return
		}
	default: // 0: standard
		labelFile = "temp/label.png"
		labelSize = STANDARD_LABEL_SIZE
		printer = STANDARD_PRINTER
		if err := formatLabel(request.ItemId, request.Serial, request.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Ok: false, Error: err.Error()})
			return
		}
	}

	args := []string{"-p", printer, "-m", BROTHER_QL_MODEL, "print", "-l", labelSize}
	for i := 0; i < request.Quantity; i++ {
		args = append(args, labelFile)
	}
	cmd := exec.Command("brother_ql", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("print failed: %s", err)
		if detail := strings.TrimSpace(string(out)); detail != "" {
			msg += ": " + detail
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Ok: false, Error: msg})
		return
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Ok:     true,
		ItemId: request.ItemId,
	})
}

func main() {
	http.HandleFunc("/printer", printHandler)

	fmt.Println("Server starting on http://localhost:6767")

	if err := http.ListenAndServe(":6767", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
