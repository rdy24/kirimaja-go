package qrcode

import (
	"fmt"
	"path/filepath"

	qr "github.com/skip2/go-qrcode"
)

type Service struct {
	publicDir string
}

func New(publicDir string) *Service {
	return &Service{publicDir}
}

// Generate saves a QR code PNG for the given content and returns the web path.
func (s *Service) Generate(content string) (string, error) {
	filename := fmt.Sprintf("%s.png", content)
	savePath := filepath.Join(s.publicDir, "qrcodes", filename)

	if err := qr.WriteFile(content, qr.Medium, 256, savePath); err != nil {
		return "", fmt.Errorf("qrcode generate failed: %w", err)
	}
	return "/qrcodes/" + filename, nil
}
