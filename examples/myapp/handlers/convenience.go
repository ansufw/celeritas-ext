package handlers

import (
	"context"
	"net/http"

	"github.com/ansufw/celeritas"
)

func (h *Handlers) render(w http.ResponseWriter, r *http.Request, template string, vars, data any) error {
	return h.App.Render.Page(w, r, template, vars, data)
}

func (h *Handlers) sessionPut(ctx context.Context, key, value string) {
	h.App.Session.Put(ctx, key, value)
}

func (h *Handlers) sessionHas(ctx context.Context, key string) bool {
	return h.App.Session.Exists(ctx, key)
}

func (h *Handlers) sessionGet(ctx context.Context, key string) any {
	return h.App.Session.Get(ctx, key)
}

func (h *Handlers) sessionRemove(ctx context.Context, key string) {
	h.App.Session.Remove(ctx, key)
}

func (h *Handlers) sessionRenew(ctx context.Context) {
	h.App.Session.RenewToken(ctx)
}

func (h *Handlers) sessionDestroy(ctx context.Context) {
	h.App.Session.Destroy(ctx)
}

func (h *Handlers) randomString(n int) string {
	return h.App.RandomString(n)
}

func (h *Handlers) encrypt(text string) (string, error) {
	enc := celeritas.Encryption{
		Key: []byte(h.App.EncryptionKey),
	}

	encrypted, err := enc.Encrypt(text)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

func (h *Handlers) decrypt(cryptedText string) (string, error) {
	enc := celeritas.Encryption{
		Key: []byte(h.App.EncryptionKey),
	}

	decrypted, err := enc.Decrypt(cryptedText)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}
