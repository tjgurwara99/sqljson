package main

import (
	"embed"
	"encoding/json"
	"hash/fnv"
	"io"
	"os"
	"text/template"

	"github.com/tjgurwara99/sqljson/internal/relationships"
)

//go:embed templates/*
var content embed.FS

var colors = map[int][]string{
	100: {
		"#FFCDD2",
		"#F8BBD0",
		"#E1BEE7",
		"#D1C4E9",
		"#C5CAE9",
		"#BBDEFB",
		"#B3E5FC",
		"#B2EBF2",
		"#B2DFDB",
		"#C8E6C9",
		"#DCEDC8",
		"#F0F4C3",
		"#FFF9C4",
		"#FFECB3",
		"#FFE0B2",
		"#FFCCBC",
		"#D7CCC8",
		"#F5F5F5",
		"#CFD8DC",
	},
	300: {
		"#E57373",
		"#F06292",
		"#BA68C8",
		"#9575CD",
		"#7986CB",
		"#64B5F6",
		"#4FC3F7",
		"#4DD0E1",
		"#4DB6AC",
		"#81C784",
		"#AED581",
		"#DCE775",
		"#FFF176",
		"#FFD54F",
		"#FFB74D",
		"#FF8A65",
		"#A1887F",
		"#E0E0E0",
		"#90A4AE",
	},
}

func Hash(key string) (int, error) {
	h := fnv.New32a()
	if _, err := h.Write([]byte(key)); err != nil {
		return 0, err
	}
	return int(h.Sum32()), nil
}

func Color(level int, index int) string {
	items := colors[level]
	return items[index%len(items)]
}

func Transform(r io.Reader, w io.Writer) error {
	dot, err := template.New("dot").Funcs(template.FuncMap{
		"Color":  Color,
		"Hash":   Hash,
		"Escape": template.HTMLEscapeString,
	}).ParseFS(content, "templates/*")
	if err != nil {
		return err
	}

	var createTables map[string]*relationships.CreateTable

	dec := json.NewDecoder(r)
	if err := dec.Decode(&createTables); err != nil {
		return err
	}

	return dot.ExecuteTemplate(w, "dot.tmpl", &createTables)
}

func main() {
	err := Transform(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}
