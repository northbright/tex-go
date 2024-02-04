package tex

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/northbright/pathelper"
)

const (
	Ext = ".tex"
)

var (
	// Show xelatex output or not.
	DebugMode = false

	// Left delimiter for tex template(default: "\{\{").
	LeftDelimter = "\\{\\{"

	// Right delimiter for tex template(default: "\}\}").
	RightDelimter = "\\}\\}"
)

func LoadTemplates(dir string) (map[string]*template.Template, error) {
	m := make(map[string]*template.Template)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) != Ext {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// Create a new template.
		t := template.New(filepath.Base(path))
		log.Printf("1: t.Templates(): %v", t.Templates())

		// Set delimiters before parsing.
		// The default Golang delimiters: "{{" and "}}"
		// conflict with LaTex which uses "{}" for command parameters.
		t = t.Delims(LeftDelimter, RightDelimter)

		log.Printf("absPath: %v", absPath)
		// Parse template file.
		t, err = t.ParseFiles(absPath)
		if err != nil {
			log.Printf("ParseFiles() error: %v", err)
			return err
		}
		log.Printf("2: t.Templates(): %v", t.Templates())
		for _, tmpl := range t.Templates() {
			log.Printf("name: %v", tmpl.Name())
		}

		m[path] = t

		return nil
	})

	if err != nil {
		return nil, err
	}

	return m, nil
}

func OutputTex(t *template.Template, outputTex string, data any) error {
	outputDir := filepath.Dir(outputTex)

	// Create output dir if it does not exists.
	if err := pathelper.CreateDirIfNotExists(outputDir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(outputTex, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	if err = t.Execute(f, data); err != nil {
		log.Printf("Execute() error: %v", err)
		return err
	}

	return nil
}

func Output(m map[string]*template.Template, outputDir string, data any) error {
	// Create output dir if it does not exists.
	if err := pathelper.CreateDirIfNotExists(outputDir, 0755); err != nil {
		return err
	}

	for path, tmpl := range m {
		// Make dir for output tex file if need.
		dir := filepath.Join(outputDir, filepath.Dir(path))
		if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
			return err
		}

		file := filepath.Join(outputDir, path)

		// Write data to the tex file.
		if err := OutputTex(tmpl, file, data); err != nil {
			log.Printf("OutputTex() error, tmpl: %v, file: %v", tmpl, file)
			return err
		}
	}

	return nil
}

/*
func RenderTemplates(dstDir string, srcDir string, data any) error {
		// Check if src templates dir exists.
		if !(pathelper.FileExists(srcDir)) {
			return fmt.Errorf("source templates dir not found")
		}

		// Set dst(output) dir to "SRC_DIR/output" if it's the same as src dir or empty.
		if dstDir == srcDir || dstDir == "" {
			dstDir = filepath.Join(srcDir, "output")
		}

		// Create dst dir if it does not exists.
		if err := pathelper.CreateDirIfNotExists(dstDir, 0755); err != nil {
			return err
		}

		// Prepare a pattern.
		pattern := filepath.Join(srcDir, "*.tex")

		// Loading a set of templates from the template directory.
		// t is the template which its file name is the first one matched the pattern.
		// e.g. "00-title.tex".
		t, err := template.ParseGlob(pattern)
		if err != nil {
			return err
		}

		// Get all templates.
		templates := t.Templates()
		return templates
	return nil
}
*/

// ToPDF compiles a tex file into the PDF file by running xelatex.
// It outputs the pdf under the source tex file's dir and returns the compiled PDF path.
func ToPDF(texFile string) (string, error) {
	// Get absolute path of tex file.
	texFileAbsPath, err := filepath.Abs(texFile)
	if err != nil {
		return "", err
	}

	// Get source tex file's dir.
	srcDir := filepath.Dir(texFileAbsPath)

	// Check if xelatex command exists.
	if !pathelper.CommandExists("xelatex") {
		return "", fmt.Errorf("xelatex does not exists")
	}

	// Run "xelatex" command to compile a tex file into a PDF under src dir 2 times.
	// 1st time: create a PDF and .aux files(cross-references) and a .toc(Table of Content).
	// 2nd time: re-create the PDF with crosss-references and TOC.
	for i := 0; i < 2; i++ {
		// Run xelatex with options:
		// -synctex=1
		// -interaction=nonstopmode
		// -shell-escape
		cmd := exec.Command("xelatex", "-synctex", "1", "-interaction", "nonstopmode", "-shell-escape", texFileAbsPath)
		// Set work dir to source tex file's dir.
		cmd.Dir = srcDir

		// Show xelatex output for DEBUG.
		if DebugMode {
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
		}

		// Run xelatex
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}

	// Get output PDF file path.
	baseFile := pathelper.BaseWithoutExt(texFile)
	pdf := filepath.Join(srcDir, baseFile+".pdf")

	// Check if PDF exists.
	if !pathelper.FileExists(pdf) {
		return "", fmt.Errorf("xelatex compiled successfully but no output pdf found")
	}

	return pdf, nil
}
