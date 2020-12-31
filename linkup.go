// LinkUp - A tool for catching broken website links.
// Copyright (C) 2020-2021 Henry G. Stratmann III
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package linkup

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

type fsEntity struct {
	name      string
	fullname  string
	directory bool
	children  map[string]*fsEntity
	parent    *fsEntity
	ids       map[string]int
	hrefs     []string
}

// Website represents a set of related web pages located under a single domain.
// Each web page can cantain zero or more links.
type Website struct {
	root *fsEntity
}

// New allocates and initializes a new instance of the Website structure.
func New() *Website {
	ent := allocateFSEntity("/")
	ent.directory = true
	return &Website{root: ent}
}

// AddFile registers a non-HTML file.
// The file could be an image, font, stylesheet, or other file.
// Its name must be relative to the root of the domain.
func (w *Website) AddFile(name string) error {
	name = prepareFileName(name)
	if newFSEntity(w.root, name) == nil {
		return fmt.Errorf("file already registered with name '%s'", name)
	}
	return nil
}

// AddDocument registers the specified file as an HTML document.
// The file name must be relative to the root of the domain.
func (w *Website) AddDocument(name string) error {
	name = prepareFileName(name)
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()
	return w.AddDocumentFromReader(name, file)
}

// AddDocumentFromReader registers the specified web page for link verification.
// The file name must be relative to the root of the domain.
func (w *Website) AddDocumentFromReader(name string, reader io.Reader) error {
	name = prepareFileName(name)
	entity := newFSEntity(w.root, name)
	if entity == nil {
		return fmt.Errorf("file already registered with name '%s'", name)
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Recursively collect all links.
	var visitNode func(i int, s *goquery.Selection)

	visitNode = func(i int, s *goquery.Selection) {
		switch strings.ToLower(goquery.NodeName(s)) {
		case "a", "link":
			if href, exists := s.Attr("href"); exists {
				entity.hrefs = append(entity.hrefs, href)
			}
			break

		case "script", "img", "source":
			if src, exists := s.Attr("src"); exists {
				entity.hrefs = append(entity.hrefs, src)
			}
			if srcsets, exists := s.Attr("srcset"); exists {
				images := strings.Split(srcsets, ",")
				for _, image := range images {
					index := strings.LastIndex(image, " ")
					if index < 0 {
						entity.hrefs = append(entity.hrefs, image)
					} else {
						entity.hrefs = append(entity.hrefs, image[:index])
					}
				}
			}
			break
		}

		if id, exists := s.Attr("id"); exists {
			entity.ids[id]++
		}

		s.Children().Each(visitNode)
	}

	doc.Each(visitNode)
	return nil
}

// Validate detects broken website links.
// All files must be registered before calling this method.
func (w *Website) Validate() []error {
	return validate(w, w.root)
}

func isPathValid(entity *fsEntity, components []string) *fsEntity {
	if entity == nil {
		return nil
	}

	if len(components) == 0 {
		if entity.directory {
			// A directory can be linked to if it contains an index file.
			for _, index := range []string{"index.html", "index.htm", "index.tmpl"} {
				if ent, exists := entity.children[index]; exists {
					return ent
				}
			}
			return nil
		}
		return entity
	}

	if components[0] == ".." {
		return isPathValid(entity.parent, components[1:])
	}

	if child, exists := entity.children[components[0]]; exists {
		return isPathValid(child, components[1:])
	}

	return nil
}

func splitPath(path string) []string {
	components := strings.Split(path, "/")
	var pieces []string
	for _, c := range components {
		if len(c) > 0 {
			pieces = append(pieces, c)
		}
	}
	return pieces
}

func validate(website *Website, entity *fsEntity) []error {
	var errors []error

	if entity.directory {
		for _, child := range entity.children {
			errors = append(errors, validate(website, child)...)
		}
		return errors
	}

	for name, count := range entity.ids {
		if count > 1 {
			errors = append(errors, fmt.Errorf("%s: id '%s' appears %d times on the page (it should only appear once)", entity.fullname, name, count))
		}
	}

	for _, href := range entity.hrefs {
		// Perform some sanitization on the string.
		href = strings.TrimSpace(href)
		href = strings.Replace(href, "\\", "/", -1)

		// Check if this is a website URL.
		if strings.HasPrefix(href, "http") {
			// Ping the URL and make sure it's active.
			status, err := ping(href)
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: encountered error when pinging '%s'", entity.fullname, href))
			} else if status != 200 {
				errors = append(errors, fmt.Errorf("%s: encountered status code %d when pinging '%s'", entity.fullname, status, href))
			}
			continue
		}

		if href == "#" {
			errors = append(errors, fmt.Errorf("%s: incomplete target '#'", entity.fullname))
			continue
		}

		if href == "/" {
			continue
		}

		hashIndex := strings.LastIndex(href, "#")
		if hashIndex == 0 {
			_, i := utf8.DecodeRuneInString(href)
			target := href[i:]
			if _, exists := entity.ids[target]; !exists {
				errors = append(errors, fmt.Errorf("%s: broken same page link '%s'", entity.fullname, href))
			}
			continue
		}

		var targetEnt *fsEntity = nil
		target := ""
		if hashIndex > 0 {
			target = strings.TrimSpace(href[hashIndex+1:])
			href = strings.TrimSpace(href[:hashIndex])
		}

		if strings.HasPrefix(href, "/") {
			if targetEnt = isPathValid(website.root, splitPath(href)); targetEnt == nil {
				errors = append(errors, fmt.Errorf("%s: broken link '%s'", entity.fullname, href))
				continue
			}
		} else {
			if targetEnt = isPathValid(entity.parent, splitPath(href)); targetEnt == nil {
				errors = append(errors, fmt.Errorf("%s: broken relative link '%s'", entity.fullname, href))
				continue
			}
		}

		if hashIndex > 0 {
			if _, exists := targetEnt.ids[target]; !exists {
				errors = append(errors, fmt.Errorf("%s: broken target link '%s#%s'", entity.fullname, href, target))
			}
		}
	}

	return errors
}

func prepareFileName(name string) string {
	// Strip away any leading slash since all files should be relative to the root.
	if strings.HasPrefix(name, "/") {
		_, i := utf8.DecodeRuneInString(name)
		return name[i:]
	}
	return name
}

func allocateFSEntity(name string) *fsEntity {
	return &fsEntity{
		name:     name,
		ids:      make(map[string]int),
		children: make(map[string]*fsEntity),
	}
}

func calcFullName(entity *fsEntity) string {
	current := entity
	fullname := ""
	for current != nil && current.parent != nil {
		if len(fullname) == 0 {
			fullname = current.name
		} else {
			fullname = current.name + "/" + fullname
		}
		current = current.parent
	}
	return fullname
}

func createFSEntity(parent *fsEntity, components []string) *fsEntity {
	if len(components) == 0 {
		return parent
	}

	if parent.directory {
		if child, exists := parent.children[components[0]]; exists {
			return createFSEntity(child, components[1:])
		}

		child := allocateFSEntity(components[0])
		child.parent = parent
		child.fullname = calcFullName(child)
		parent.children[components[0]] = child

		if len(components) > 1 {
			child.directory = true
			return createFSEntity(child, components[1:])
		}
		return child
	}

	// A file already exists with this name.
	// Don't create a duplicate.
	return nil
}

func newFSEntity(root *fsEntity, path string) *fsEntity {
	return createFSEntity(root, strings.Split(path, "/"))
}

func ping(url string) (int, error) {
	var client = http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}
