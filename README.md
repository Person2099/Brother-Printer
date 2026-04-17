# Label Printer

Go-based HTTP service for generating and printing labels with QR codes for Monash Automation inventory items.

## Features

- Generates labels with item serial numbers, names, and QR codes
- Includes Monash Automation branding
- Prints to Brother QL-700 label printer via CUPS
- Supports multiple label types: standard, small, and cable
- Routes to the appropriate printer and label stock automatically
- REST API for label printing

## Prerequisites

- Go 1.16+
- Brother QL-700 label printers configured in CUPS (one for standard/cable, one for small)
- CUPS printing system (Linux/macOS)

## Installation

1. **Clone the repository:**
```bash
git clone https://github.com/yourusername/Label-Printer.git
cd Label-Printer
```

2. **Install dependencies:**
```bash
go mod download  # Downloads dependencies
go install       # Compiles and installs binary
```

3. **Configure printers:**
   - Ensure the standard/cable printer is installed in CUPS as `brother_ql.700`
   - Ensure the small label printer is installed in CUPS as `brother_ql.700_small`
   - Verify with: `lpstat -p -d`
   - To use different CUPS names, update `STANDARD_PRINTER_NAME` and `SMALL_PRINTER_NAME` in `main.go`

## Usage

### Starting the Server

```bash
go run .
```

Server starts on `http://localhost:6767`

### API Endpoint

**POST** `/printer`

Generates and prints a label.

**Request Body:**
```json
{
  "name": "Arduino Uno R3",
  "serial": "SN12345",
  "quantity": 2,
  "itemId": "550e8400-e29b-41d4-a716-446655440000",
  "labelType": 0
}
```

**Parameters:**
- `name` (string, required): Item name displayed on label
- `serial` (string, required): Serial number displayed prominently
- `quantity` (int, required): Number of labels to print (must be > 0)
- `itemId` (string, required): Valid UUID encoded in QR code
- `labelType` (int, optional): Label format — `0` standard (default), `1` small, `2` cable

**Success Response:**
```json
{
  "ok": true,
  "itemId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Response:**
```json
{
  "ok": false,
  "error": "Error description"
}
```

### Example Requests

**Standard label (DK-11201, 29mm × 90mm):**
```bash
curl -X POST http://localhost:6767/printer \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Raspberry Pi 4",
    "serial": "RPI-001",
    "quantity": 1,
    "itemId": "123e4567-e89b-12d3-a456-426614174000",
    "labelType": 0
  }'
```

**Small label (DK-11204, 17mm × 54mm):**
```bash
curl -X POST http://localhost:6767/printer \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Resistor 10kΩ",
    "serial": "RES-042",
    "quantity": 5,
    "itemId": "123e4567-e89b-12d3-a456-426614174000",
    "labelType": 1
  }'
```

**Cable label (DK-11201, 29mm × 90mm — same format as standard):**
```bash
curl -X POST http://localhost:6767/printer \
  -H "Content-Type: application/json" \
  -d '{
    "name": "USB-A to USB-C",
    "serial": "CAB-007",
    "quantity": 1,
    "itemId": "123e4567-e89b-12d3-a456-426614174000",
    "labelType": 2
  }'
```

## Label Specifications

| Type | Stock | Physical Size | Canvas | Printer |
|------|-------|--------------|--------|---------|
| Standard (0) | DK-11201 | 29mm × 90mm | 306 × 991 px | `brother_ql.700` |
| Small (1) | DK-11204 | 17mm × 54mm | 187 × 594 px | `brother_ql.700_small` |
| Cable (2) | DK-11201 | 29mm × 90mm | 306 × 991 px | `brother_ql.700` |

**Standard / Cable label:**
- QR code: 241 × 241 px (High error correction)
- Margins: 30 px
- Serial: 100pt bold · Name: 40pt regular · Header: 30pt regular

**Small label:**
- QR code: 160 × 160 px (High error correction)
- Margins: 10 px
- Serial: 52pt bold · Name: 20pt regular

## Project Structure

```
Label-Printer/
├── main.go           # HTTP server and print handler
├── validation.go     # Request validation
├── format.go         # Label generation logic
├── assets/           # Logo images
├── fonts/            # Custom fonts
└── temp/             # Generated label output
```

## Dependencies

- `github.com/golang/freetype` - Font rendering
- `github.com/nfnt/resize` - Image resizing
- `github.com/skip2/go-qrcode` - QR code generation
- `github.com/google/uuid` - UUID validation
- `golang.org/x/image/font/gofont/*` - Embedded fallback fonts

## License

[Add your license here]
