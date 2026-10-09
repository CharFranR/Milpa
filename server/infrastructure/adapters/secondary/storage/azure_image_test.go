package storage

import "testing"

func TestBuildPublicImageURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		filename string
		want     string
	}{
		{
			name:     "base sin slash final",
			base:     "https://milpa.example",
			filename: "img-1.jpg",
			want:     "https://milpa.example/api/v1/images/img-1.jpg",
		},
		{
			name:     "base con slash final",
			base:     "https://milpa.example/",
			filename: "img-1.jpg",
			want:     "https://milpa.example/api/v1/images/img-1.jpg",
		},
		{
			name:     "base con varios slash finales",
			base:     "https://milpa.example///",
			filename: "img-1.jpg",
			want:     "https://milpa.example/api/v1/images/img-1.jpg",
		},
		{
			name:     "escapa el nombre del archivo",
			base:     "http://localhost:8080",
			filename: "foto 1.png",
			want:     "http://localhost:8080/api/v1/images/foto%201.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildPublicImageURL(tt.base, tt.filename); got != tt.want {
				t.Fatalf("buildPublicImageURL(%q, %q) = %q, want %q", tt.base, tt.filename, got, tt.want)
			}
		})
	}
}

func TestResolveContentType(t *testing.T) {
	pngHeader := []byte("\x89PNG\r\n\x1a\n")
	jpegHeader := []byte("\xff\xd8\xff\xe0")

	t.Run("usa la extension en minusculas y mayusculas", func(t *testing.T) {
		if got := resolveContentType("foto.PNG", nil); got != "image/png" {
			t.Fatalf("resolveContentType(foto.PNG) = %q, want image/png", got)
		}
	})

	t.Run("usa la extension jpeg", func(t *testing.T) {
		if got := resolveContentType("foto.jpeg", nil); got != "image/jpeg" {
			t.Fatalf("resolveContentType(foto.jpeg) = %q, want image/jpeg", got)
		}
	})

	t.Run("cae al sniffing cuando no hay extension conocida", func(t *testing.T) {
		if got := resolveContentType("foto-sin-extension", pngHeader); got != "image/png" {
			t.Fatalf("resolveContentType(sin extension, png) = %q, want image/png", got)
		}
		if got := resolveContentType("foto-sin-extension", jpegHeader); got != "image/jpeg" {
			t.Fatalf("resolveContentType(sin extension, jpeg) = %q, want image/jpeg", got)
		}
	})
}
