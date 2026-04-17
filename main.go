package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

const (
	STANDARD_PRINTER_NAME = "brother_ql.700"
	SMALL_PRINTER_NAME    = "brother_ql.700_small"
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
	var printerName string

	switch request.LabelType {
	case 1: // small
		labelFile = "temp/label_small.png"
		printerName = SMALL_PRINTER_NAME
		if err := formatSmallLabel(request.ItemId, request.Serial, request.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Ok: false, Error: err.Error()})
			return
		}
	default: // 0: standard, 2: cable (both use standard format)
		labelFile = "temp/label.png"
		printerName = STANDARD_PRINTER_NAME
		if err := formatLabel(request.ItemId, request.Serial, request.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Ok: false, Error: err.Error()})
			return
		}
	}

	cmd := exec.Command("lp", "-d", printerName, "-n", fmt.Sprintf("%d", request.Quantity), labelFile)
	err := cmd.Run()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Ok:    false,
			Error: err.Error(),
		})
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
