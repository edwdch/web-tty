package web

import (
	"embed"
	"io/fs"
)

// dist must contain a Vite build (index.html, assets). Empty dir fails compile.
//go:embed dist
var dist embed.FS

func Files() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
