package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	var inputFile string
	var outputDir string

	flag.StringVar(&inputFile, "input", "", "Path to the input HTML file")
	flag.StringVar(&outputDir, "output", "", "Path to the output directory")
	flag.Parse()

	if inputFile == "" || outputDir == "" {
		fmt.Println("Both -input and -output flags are required")
		flag.Usage()
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	contentBytes, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	doc, err := html.Parse(bytes.NewReader(contentBytes))
	if err != nil {
		fmt.Printf("Error parsing HTML: %v\n", err)
		os.Exit(1)
	}

	var groupItems []*html.Node
	var findGroupItems func(*html.Node)
	findGroupItems = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" {
					classes := strings.Fields(a.Val)
					for _, c := range classes {
						if c == "group-item" {
							groupItems = append(groupItems, n)
							break
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findGroupItems(c)
		}
	}
	findGroupItems(doc)

	fmt.Printf("Found %d group items.\n", len(groupItems))

	globalMethodIndex := 0

	for i, n := range groupItems {
		// Process method groups within this group item
		var methodGroups []*html.Node
		var findMethodGroups func(*html.Node)
		findMethodGroups = func(curr *html.Node) {
			if curr.Type == html.ElementNode && curr.Data == "div" {
				for _, a := range curr.Attr {
					if a.Key == "class" {
						classes := strings.Fields(a.Val)
						for _, c := range classes {
							if c == "api-method-group" {
								methodGroups = append(methodGroups, curr)
								// Don't traverse inside an api-method-group looking for more api-method-groups
								return 
							}
						}
					}
				}
			}
			for c := curr.FirstChild; c != nil; c = c.NextSibling {
				findMethodGroups(c)
			}
		}
		findMethodGroups(n)

		if len(methodGroups) > 0 {
			methodsDir := filepath.Join(outputDir, "methods")
			if err := os.MkdirAll(methodsDir, 0755); err != nil {
				fmt.Printf("Error creating methods directory: %v\n", err)
			}

			for _, mg := range methodGroups {
				// Extract ID
				mgID := ""
				for _, a := range mg.Attr {
					if a.Key == "id" {
						mgID = a.Val
						break
					}
				}
				
				if mgID == "" {
					// Try to find h1/h2 with id inside if the div itself doesn't have it (though provided snippet shows div has it)
					// Fallback
										mgID = "unknown_method"
									}
					
				globalMethodIndex++
				mgFilename := fmt.Sprintf("%02d-%s.html", globalMethodIndex, mgID)
				mgPath := filepath.Join(methodsDir, mgFilename)

				var mgBuf bytes.Buffer
				if err := html.Render(&mgBuf, mg); err != nil {
				fmt.Printf("Error rendering method group %s: %v\n", mgFilename, err)
				continue
			}

			if err := os.WriteFile(mgPath, mgBuf.Bytes(), 0644); err != nil {
				fmt.Printf("Error writing method group file %s: %v\n", mgFilename, err)
			} else {
				fmt.Printf("Wrote method group %s\n", mgFilename)
			}

			// Remove from parent
			if mg.Parent != nil {
				mg.Parent.RemoveChild(mg)
			}
		}
		}

		// Now write the group item (potentially stripped of its method groups)
		id := ""
		// Find h1 with id
		var findID func(*html.Node)
		findID = func(curr *html.Node) {
			if id != "" {
				return
			}
			if curr.Type == html.ElementNode && curr.DataAtom == atom.H1 {
				for _, a := range curr.Attr {
					if a.Key == "id" {
						id = a.Val
						return
					}
				}
			}
			for c := curr.FirstChild; c != nil; c = c.NextSibling {
				findID(c)
			}
		}
		findID(n)

		filename := ""
		if id != "" {
			filename = fmt.Sprintf("%02d-%s.html", i+1, id)
		} else {
			filename = fmt.Sprintf("%02d-chunk_%d.html", i+1, i)
		}

		var buf bytes.Buffer
		if err := html.Render(&buf, n); err != nil {
			fmt.Printf("Error rendering node %s: %v\n", filename, err)
			continue
		}

		outputPath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
			fmt.Printf("Error writing file %s: %v\n", filename, err)
		} else {
			fmt.Printf("Wrote %s\n", filename)
		}
	}
}