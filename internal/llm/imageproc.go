package llm

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/gif" // register gif decoder
	_ "image/png" // register png decoder
	"math"
	"strconv"
)

const (
	maxReceiptDim = 1280        // max pixels on longest side before resize
	skipThreshold = 700 * 1024 // skip resize if already ≤ 700 KB
	jpegQuality   = 78
	maxPDFBytes   = 3 << 20 // 3 MB hard limit for PDFs
	maxPDFPages   = 3       // reject PDFs with more than this many pages
)

// PreprocessReceipt reduces image size before sending to the LLM vision API.
//
// Images: if over 700 KB, resized to max 1280 px longest side then re-encoded
// as JPEG at quality 78. A typical 8 MB phone photo compresses to ~100–200 KB.
//
// PDFs: Anthropic processes each page as a vision image, so large PDFs are
// very expensive. PDFs over 3 MB are rejected with a user-friendly error;
// the caller should tell the user to send a photo instead.
func PreprocessReceipt(data []byte, mediaType string) ([]byte, string, error) {
	if mediaType == "application/pdf" {
		return preprocessPDF(data)
	}
	return preprocessImage(data, mediaType)
}

func preprocessPDF(data []byte) ([]byte, string, error) {
	if len(data) > maxPDFBytes {
		mb := float64(len(data)) / (1024 * 1024)
		return nil, "", fmt.Errorf(
			"PDF is too large (%.1f MB, max 3 MB). "+
				"Please send a photo of the receipt instead, or export a smaller PDF",
			mb,
		)
	}
	if pages := pdfPageCount(data); pages > maxPDFPages {
		return nil, "", fmt.Errorf(
			"PDF has %d pages (max %d). "+
				"Please send only the receipt page as a photo or a single-page PDF",
			pages, maxPDFPages,
		)
	}
	return data, "application/pdf", nil
}

// pdfPageCount extracts the total page count from a PDF without a full parser.
// Every conforming PDF stores the total in the root Pages dictionary as "/Count N".
// Returns 0 if the count cannot be determined (encrypted or non-standard PDF),
// which causes the page-count check to be skipped (fail-open).
func pdfPageCount(data []byte) int {
	needle := []byte("/Count ")
	max := 0
	pos := 0
	for pos < len(data) {
		idx := bytes.Index(data[pos:], needle)
		if idx < 0 {
			break
		}
		pos += idx + len(needle)
		// Read the digit sequence that follows "/Count "
		end := pos
		for end < len(data) && data[end] >= '0' && data[end] <= '9' {
			end++
		}
		if end > pos {
			if n, err := strconv.Atoi(string(data[pos:end])); err == nil && n > max {
				max = n
			}
		}
	}
	return max
}

func preprocessImage(data []byte, mediaType string) ([]byte, string, error) {
	if len(data) <= skipThreshold {
		return data, mediaType, nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// Unsupported format (e.g. WEBP without decoder registered) — send as-is.
		return data, mediaType, nil
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > maxReceiptDim || h > maxReceiptDim {
		img = scaleDown(img, maxReceiptDim)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return data, mediaType, nil // encoding failed — send original
	}
	return buf.Bytes(), "image/jpeg", nil
}

// scaleDown returns a new image with the longest side scaled to maxDim,
// preserving aspect ratio.
func scaleDown(src image.Image, maxDim int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	var nw, nh int
	if w >= h {
		nw = maxDim
		nh = int(math.Round(float64(h) * float64(maxDim) / float64(w)))
	} else {
		nh = maxDim
		nw = int(math.Round(float64(w) * float64(maxDim) / float64(h)))
	}
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	return resizeNearest(src, nw, nh)
}

// resizeNearest is a nearest-neighbour downscaler. Fast and sufficient for
// receipt OCR — text legibility matters, not photo quality.
func resizeNearest(src image.Image, nw, nh int) image.Image {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy := sb.Min.Y + int(float64(y)*float64(sh)/float64(nh))
		for x := 0; x < nw; x++ {
			sx := sb.Min.X + int(float64(x)*float64(sw)/float64(nw))
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
