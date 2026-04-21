package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

const (
	WIDTH      = 991
	HEIGHT     = 306
	MARGINS    = 30
	QR_CODE_LW = 241

	// DK-11204: 17mm x 54mm at ~11px/mm
	SMALL_WIDTH      = 594
	SMALL_HEIGHT     = 187
	SMALL_MARGINS    = 10
	SMALL_QR_CODE_LW = 160
)

func formatLabel(itemId, serial, name string) error {
	// Create blank white label
	canvas := image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT))
	draw.Draw(canvas, canvas.Bounds(), image.White, image.Point{}, draw.Src)

	// Add the text
	normalFont := "fonts/roboto-font/RobotoRegular.ttf"
	boldFont  := "fonts/roboto-font/RobotoBlack-Powx.ttf"

	const (
		nameFontSize   = 40.0
		nameLineHeight = 46
		titleY         = MARGINS + 15
		serialFontSize = 100.0
		headerBottom   = titleY + 44 // below 40px logo with gap
	)

	addTextWithFont(canvas, MARGINS+30+25, titleY, "Monash Automation", 30, normalFont, false)

	textAreaWidth := WIDTH - QR_CODE_LW - MARGINS*3
	nameLines := wrapText(name, textAreaWidth, nameFontSize, normalFont, false)
	nameBlockHeight := len(nameLines) * nameLineHeight
	nameStartY := HEIGHT - MARGINS - nameBlockHeight

	// Position serial between header and name block, with 8px gaps
	serialY := nameStartY - int(serialFontSize) - 8
	if serialY < headerBottom {
		serialY = headerBottom
	}
	addTextWithFont(canvas, MARGINS, serialY, serial, serialFontSize, boldFont, true)

	for i, line := range nameLines {
		addTextWithFont(canvas, MARGINS, nameStartY+i*nameLineHeight, line, nameFontSize, normalFont, false)
	}

	if err := overlayImage(canvas, "assets/monash_automation_logo.png", MARGINS, titleY, 40, 40); err != nil {
		return fmt.Errorf("failed to overlay logo: %w", err)
	}

	if err := createQR(canvas, itemId, QR_CODE_LW); err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	outFile, err := os.Create("temp/label.png")
	if err != nil {
		return fmt.Errorf("failed to create label file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, canvas); err != nil {
		return fmt.Errorf("failed to encode label image: %w", err)
	}
	return nil
}

func formatSmallLabel(itemId, serial, name string) error {
	// DK-11204: 17mm x 54mm canvas
	canvas := image.NewRGBA(image.Rect(0, 0, SMALL_WIDTH, SMALL_HEIGHT))
	draw.Draw(canvas, canvas.Bounds(), image.White, image.Point{}, draw.Src)

	boldFont := "fonts/roboto-font/RobotoBlack-Powx.ttf"
	normalFont := "fonts/roboto-font/RobotoRegular.ttf"

	// Header row: logo + "Monash Automation" — anchored to top
	const (
		logoSize       = 30
		headerFontSize = 22
		serialFontSize = 80
		nameFontSize   = 20
		nameLineHeight = 24 // nameFontSize + 4px leading
	)
	_ = overlayImage(canvas, "assets/monash_automation_logo.png", SMALL_MARGINS, SMALL_MARGINS, logoSize, logoSize)
	addTextWithFont(canvas, SMALL_MARGINS+logoSize+6, SMALL_MARGINS, "Monash Automation", headerFontSize, normalFont, false)

	// Serial number — fills the middle
	addTextWithFont(canvas, SMALL_MARGINS, SMALL_MARGINS+logoSize+4, serial, serialFontSize, boldFont, true)

	// Item name — word-wrapped, anchored to bottom
	textAreaWidth := SMALL_WIDTH - SMALL_QR_CODE_LW - SMALL_MARGINS*3
	nameLines := wrapText(name, textAreaWidth, nameFontSize, normalFont, false)
	nameBlockHeight := len(nameLines) * nameLineHeight
	nameStartY := SMALL_HEIGHT - SMALL_MARGINS - nameBlockHeight
	for i, line := range nameLines {
		addTextWithFont(canvas, SMALL_MARGINS, nameStartY+i*nameLineHeight, line, nameFontSize, normalFont, false)
	}

	qrOffset := (SMALL_HEIGHT - SMALL_QR_CODE_LW) / 2
	if qrOffset < 0 {
		qrOffset = 0
	}
	if err := createSmallQR(canvas, itemId, SMALL_QR_CODE_LW, qrOffset); err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	outFile, err := os.Create("temp/label_small.png")
	if err != nil {
		return fmt.Errorf("failed to create label file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, canvas); err != nil {
		return fmt.Errorf("failed to encode label image: %w", err)
	}
	return nil
}

func formatSmallCableLabel(itemId, serial, name string) error {
	// DK-11204 cable label: wraps around cable at centre fold.
	// 7.5mm gap each side of the fold (x=297) → text zone 0–215px, QR zone 379–594px.
	canvas := image.NewRGBA(image.Rect(0, 0, SMALL_WIDTH, SMALL_HEIGHT))
	draw.Draw(canvas, canvas.Bounds(), image.White, image.Point{}, draw.Src)

	boldFont := "fonts/roboto-font/RobotoBlack-Powx.ttf"
	normalFont := "fonts/roboto-font/RobotoRegular.ttf"

	const (
		// 7.5mm × 11px/mm ≈ 82px each side of the centre fold
		gapHalf        = 82
		foldX          = SMALL_WIDTH / 2 // 297px
		textZoneW      = foldX - gapHalf // 215px
		qrZoneX        = foldX + gapHalf // 379px
		logoSize       = 20
		headerFontSize = 14.0
		serialFontSize = 46.0
		nameFontSize   = 17.0
		nameLineHeight = 21
	)

	textAreaW := textZoneW - SMALL_MARGINS*2

	// Header row: logo + "Monash Automation"
	_ = overlayImage(canvas, "assets/monash_automation_logo.png", SMALL_MARGINS, SMALL_MARGINS, logoSize, logoSize)
	addTextWithFont(canvas, SMALL_MARGINS+logoSize+4, SMALL_MARGINS, "Monash Automation", headerFontSize, normalFont, false)

	// Serial — below header
	addTextWithFont(canvas, SMALL_MARGINS, SMALL_MARGINS+logoSize+2, serial, serialFontSize, boldFont, true)

	// Name — word-wrapped, pinned to bottom of left zone
	nameLines := wrapText(name, textAreaW, nameFontSize, normalFont, false)
	nameBlockH := len(nameLines) * nameLineHeight
	nameStartY := SMALL_HEIGHT - SMALL_MARGINS - nameBlockH
	for i, line := range nameLines {
		addTextWithFont(canvas, SMALL_MARGINS, nameStartY+i*nameLineHeight, line, nameFontSize, normalFont, false)
	}

	// QR — centred in the right zone
	qrZoneW := SMALL_WIDTH - qrZoneX
	qrSize := SMALL_HEIGHT - SMALL_MARGINS*2 // 167px
	if qrSize > qrZoneW-SMALL_MARGINS*2 {
		qrSize = qrZoneW - SMALL_MARGINS*2
	}
	qrX := qrZoneX + (qrZoneW-qrSize)/2
	qrY := (SMALL_HEIGHT - qrSize) / 2
	if err := createQRAt(canvas, itemId, qrSize, qrX, qrY); err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	outFile, err := os.Create("temp/label_small_cable.png")
	if err != nil {
		return fmt.Errorf("failed to create label file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, canvas); err != nil {
		return fmt.Errorf("failed to encode label image: %w", err)
	}
	return nil
}

func createQRAt(canvas *image.RGBA, itemId string, size, x, y int) error {
	qr, err := qrcode.New(QR_HEAD_URL+itemId, qrcode.High)
	if err != nil {
		return err
	}
	qr.DisableBorder = true
	qrImg := qr.Image(size)
	offset := image.Pt(x, y)
	draw.Draw(canvas, qrImg.Bounds().Add(offset), qrImg, image.Point{}, draw.Over)
	return nil
}

func createSmallQR(canvas *image.RGBA, itemId string, length, yOffset int) error {
	qr, err := qrcode.New(QR_HEAD_URL+itemId, qrcode.High)
	if err != nil {
		return err
	}

	qr.DisableBorder = true
	qrImg := qr.Image(length)

	offset := image.Pt(SMALL_WIDTH-length-SMALL_MARGINS, yOffset)
	draw.Draw(canvas, qrImg.Bounds().Add(offset), qrImg, image.Point{}, draw.Over)
	return nil
}

func createQR(canvas *image.RGBA, itemId string, length int) error {
	qr, err := qrcode.New(QR_HEAD_URL+itemId, qrcode.High)
	if err != nil {
		return err
	}

	qr.DisableBorder = true
	qrImg := qr.Image(length)

	// Overlay on canvas
	offset := image.Pt(WIDTH-length-MARGINS, MARGINS)
	draw.Draw(canvas, qrImg.Bounds().Add(offset), qrImg, image.Point{}, draw.Over)
	return nil
}

func overlayImage(canvas *image.RGBA, imagePath string, x, y, size_x, size_y int) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image %q: %w", imagePath, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image %q: %w", imagePath, err)
	}
	img = resize.Resize(uint(size_x), uint(size_y), img, resize.Lanczos3)

	offset := image.Pt(x, y)
	draw.Draw(canvas, img.Bounds().Add(offset), img, image.Point{}, draw.Over)
	return nil
}

func wrapText(text string, maxWidth int, fontSize float64, fontPath string, bold bool) []string {
	fontData := loadFontData(fontPath, bold)
	f, err := truetype.Parse(fontData)
	if err != nil {
		return []string{text}
	}
	face := truetype.NewFace(f, &truetype.Options{Size: fontSize, DPI: 72})
	defer face.Close()

	words := strings.Fields(text)
	var lines []string
	current := ""

	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if measureString(candidate, face) > maxWidth && current != "" {
			lines = append(lines, current)
			current = word
		} else {
			current = candidate
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func measureString(s string, face xfont.Face) int {
	width := fixed.Int26_6(0)
	var prev rune
	for i, r := range s {
		if i > 0 {
			width += face.Kern(prev, r)
		}
		adv, _ := face.GlyphAdvance(r)
		width += adv
		prev = r
	}
	return int(width >> 6)
}

func loadFontData(fontPath string, bold bool) []byte {
	if fontPath != "" {
		data, err := os.ReadFile(fontPath)
		if err == nil {
			return data
		}
	}
	if bold {
		return gomonobold.TTF
	}
	return goregular.TTF
}

func addTextWithFont(img *image.RGBA, x, y int, text string, fontSize float64, fontPath string, bold bool) {
	col := color.RGBA{0, 0, 0, 255}
	
	var fontData []byte
	if fontPath != "" {
		data, err := os.ReadFile(fontPath)
		if err == nil {
			fontData = data
		}
	}
	
	// Fallback to embedded fonts
	if fontData == nil {
		if bold {
			fontData = gomonobold.TTF
		} else {
			fontData = goregular.TTF
		}
	}
	
	f, _ := truetype.Parse(fontData)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(col))
	pt := freetype.Pt(x, y+int(c.PointToFixed(fontSize)>>6))
	c.DrawString(text, pt)
}
