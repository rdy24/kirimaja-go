package qrcode

import (
	"bytes"
	"context"
	"fmt"

	qr "github.com/skip2/go-qrcode"

	"kirimaja-go/internal/common/storage"
)

type Service struct {
	store storage.Store
}

func New(store storage.Store) *Service {
	return &Service{store: store}
}

// Generate renders a QR PNG in memory and stores it in object storage,
// returning the object key (persisted on the shipment; resolved to a
// presigned URL when served).
func (s *Service) Generate(ctx context.Context, content string) (string, error) {
	png, err := qr.Encode(content, qr.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("qrcode encode failed: %w", err)
	}
	key := "qrcodes/" + content + ".png"
	if err := s.store.Put(ctx, key, bytes.NewReader(png), int64(len(png)), "image/png"); err != nil {
		return "", err
	}
	return key, nil
}
