package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
)

type Request struct {
	Name      string `json:"name"`
	Serial    string `json:"serial"`
	Quantity  int    `json:"quantity"`
	ItemId    string `json:"itemId"`
	LabelType int    `json:"labelType"` // 0: standard, 1: small, 2: small cable
}

type SuccessResponse struct {
	Ok     bool   `json:"ok"`
	ItemId string `json:"itemId"`
}

type ErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// Check for malformed JSON Requests
func malformedJson(req *Request, write http.ResponseWriter, read *http.Request) bool {
	err := json.NewDecoder(read.Body).Decode(&req)
	if err != nil {
		write.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(write).Encode(ErrorResponse{
			Ok:    false,
			Error: "Json is malformed",
		})
		return true
	}
	return false
}

// Validate the inputs
func validJson(req *Request, write http.ResponseWriter) bool {
	if req.Name == "" {
		write.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(write).Encode(ErrorResponse{
			Ok:    false,
			Error: "Name is required",
		})
		return false
	}

	if req.Quantity <= 0 {
		write.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(write).Encode(ErrorResponse{
			Ok:    false,
			Error: "Quantity must be greater than 0",
		})
		return false
	}

	if req.Serial == "" {
		write.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(write).Encode(ErrorResponse{
			Ok:    false,
			Error: "Serial is required",
		})
		return false
	}

	if _, err := uuid.Parse(req.ItemId); err != nil {
		write.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(write).Encode(ErrorResponse{
			Ok:    false,
			Error: "ItemId must be a valid UUID",
		})
		return false
	}

	if req.LabelType < 0 || req.LabelType > 2 {
		write.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(write).Encode(ErrorResponse{
			Ok:    false,
			Error: "LabelType must be 0 (standard), 1 (small), or 2 (small cable)",
		})
		return false
	}

	return true
}
