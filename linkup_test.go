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
	"os"
	"path/filepath"
	"testing"
)

func TestLinksFromRoot(t *testing.T) {
	w := New()
	addWebsite("testdata/absolute", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{})
}

func TestInvalidLinksFromRoot(t *testing.T) {
	w := New()
	addWebsite("testdata/absolute_error", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"blog/first-post.html: broken link '/blog/second-post.html'",
		"blog/index.html: broken link '/home.html'",
		"blog/index.html: broken link '/first-post.html'",
	})
}

func TestRelativeLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/relative", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{})
}

func TestInvalidRelativeLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/relative_error", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"blog/index.html: broken relative link '../../index.html'",
		"blog/index.html: broken relative link '../blog/second-post.html'",
		"index.html: broken relative link 'download/../index.html'",
	})
}

func TestExternalLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/external", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{})
}

func TestInvalidExternalLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/external_error", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: encountered status code 404 when pinging 'https://www.google.com/does_not_exist'",
		"index.html: encountered error when pinging 'https://fake12371ivnd985Vkf8K98Qnm.com/'",
	})
}

func TestTargetLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/target", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{})
}

func TestInvalidTargetLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/target_error", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: broken same page link '#goodbye-world'",
		"index.html: broken same page link '#FOO'",
		"index.html: incomplete target '#'",
		"blog/index.html: broken target link '/#'",
		"blog/index.html: broken target link '../#razzle'",
		"blog/index.html: broken target link '/index.html#dazzle'",
	})
}

func TestDirectoryLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/directory", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{})
}

func TestInvalidDirectoryLinks(t *testing.T) {
	w := New()
	addWebsite("testdata/directory_error", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"blog/index.html: broken relative link '../blog/post'",
		"blog/index.html: broken relative link 'post'",
		"blog/index.html: broken link '/blog/post'",
		"index.html: broken relative link 'download'",
		"index.html: broken link '/download'",
		"index.html: broken relative link 'blog/post'",
		"index.html: broken link '/blog/post'",
	})
}

func TestInvalidStylesheets(t *testing.T) {
	w := New()
	addWebsite("testdata/link_tag", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: broken relative link 'fake.css'",
	})
}

func TestInvalidJavaScript(t *testing.T) {
	w := New()
	addWebsite("testdata/script_tag", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: broken relative link 'fake.js'",
	})
}

func TestInvalidImage(t *testing.T) {
	w := New()
	addWebsite("testdata/img_tag", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: broken relative link 'frown.png'",
	})
}

func TestInvalidImageResponsive(t *testing.T) {
	w := New()
	addWebsite("testdata/img_srcset_tag", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: broken relative link 'frown.png'",
		"index.html: broken relative link 'frown-2x.png'",
		"index.html: broken relative link 'frown-4x.png'",
	})
}

func TestInvalidSource(t *testing.T) {
	w := New()
	addWebsite("testdata/source_tag", w)
	errs := w.Validate()
	verifyErrors(t, errs, []string{
		"index.html: broken relative link 'upset.png'",
		"index.html: broken relative link 'upset-2x.png'",
		"index.html: broken relative link 'upset-4x.png'",
	})
}

func verifyErrors(t *testing.T, actualErrors []error, expectedErrors []string) {
	if len(actualErrors) != len(expectedErrors) {
		t.Error("Error count mismatch", len(actualErrors), len(expectedErrors))
	}

	// Verify all errors were expected.
	for _, actualError := range actualErrors {
		match := false
		for _, expectedErrors := range expectedErrors {
			// Check for a match.
			if actualError.Error() == expectedErrors {
				match = true
				break
			}
		}
		if !match {
			t.Error("Unexpected error", actualError)
		}
	}
}

func addWebsite(path string, website *Website) {
	// Change the current working directory.
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(path)
	if err != nil {
		panic(err)
	}

	// Treat the directory as the root of the website.
	newDir, _ := os.Getwd()
	filepath.Walk(newDir,
		func(name string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			relativeName := name[len(newDir):]
			switch filepath.Ext(relativeName) {
			case ".html", ".htm", ".tmpl":
				return website.AddDocument(relativeName)
			default:
				return website.AddFile(relativeName)
			}
		})

	// Restore the original current working directory
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func printWebsiteFiles(entity *fsEntity, depth int) {
	for i := 0; i < depth*4; i++ {
		print(" ")
	}

	if entity.directory && entity.parent != nil {
		print("/")
	}

	print(entity.name)
	print(" ")
	println(entity.fullname)

	if entity.directory {
		for _, child := range entity.children {
			printWebsiteFiles(child, depth+1)
		}
	}
}
