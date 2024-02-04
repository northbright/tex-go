package tex_test

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/northbright/tex-go"
)

func ExampleToPDF() {
	// Open DEBUG mode if need.
	tex.DebugMode = true

	texFile := "templates/manual.tex"

	// Compile a tex file to PDF.
	pdf, err := tex.ToPDF(texFile)
	if err != nil {
		log.Printf("ToPDF() error: %v", err)
		return
	}

	fmt.Printf("ToPDF OK, output pdf: %v\n", filepath.Base(pdf))

	// Output:
	//ToPDF OK, output pdf: my_book.pdf
}

func ExampleLoadTemplates() {
	// Dependency represents the dependency of tex-go.
	type Dependency struct {
		Name string
		Desc string
		URL  string
	}

	// Manual represents the manual of tex-go.
	type Manual struct {
		Title        string
		Author       string
		About        string
		Dependencies []Dependency
		ExampleCode  string
	}

	manual := Manual{
		Title:  "tex-go Manual",
		Author: "Frank Xu",
		About:  "A Go library provides Latex utilities like rendering LaTex templates and compiling a Tex file to a PDF.",
		Dependencies: []Dependency{
			Dependency{
				Name: "TexLive",
				Desc: "tex2pdf calls `xelatex` command which comes with installation of TexLive.",
				URL:  "https://tug.org/texlive/",
			},
			Dependency{
				Name: "minted",
				Desc: "minted is used for code highlighting. TexLive Installation with scheme-full includes minted.",
				URL:  "https://www.ctan.org/pkg/minted",
			},
			Dependency{
				Name: "pygments",
				Desc: "pygments is required by minted.",
				URL:  "https://www.ctan.org/pkg/minted",
			},
		},
		ExampleCode: `
		package main

		import (
                    "github.com/northbright/tex-go"
		)
		`,
	}

	srcDir := "templates"
	dstDir := "out"

	// Load templates.
	m, err := tex.LoadTemplates(srcDir)
	if err != nil {
		log.Printf("LoadTemplates() error: %v", err)
		return
	}

	// Output
	if err = tex.Output(m, dstDir, manual); err != nil {
		log.Printf("Output() error: %v", err)
		return
	}

	// Output:
}
