package util

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

// PrintQR writes a compact terminal QR code for url to stdout.
func PrintQR(url string) error {
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return fmt.Errorf("qr: %w", err)
	}
	// ToSmallString uses half-block characters — dense and readable.
	fmt.Println(q.ToSmallString(false))
	return nil
}
