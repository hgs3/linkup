# Website Link Checker

**LinkUp** is a Go package for detecting broken links on your website.

It works by inspecting your sites HTML documents and verifying all links refer to a valid location.

It understands links in `a`, `img`, `script`, `link`, and `source` tags.
External links are verified by pinging them.

[![Actions Status](https://github.com/hgs3/linkup/workflows/Build%20Status/badge.svg)](https://github.com/hgs3/linkup/actions)

## Example Usage

All documents and files are added relative to the root of the domain.

```go
w := linkup.New()
w.AddDocument("index.html")
w.AddFile("img/hello.png")
for _, err := range w.Validate() {
    println(err) // Report broken links.
}
```

## Installation

LinkUp depends upon the [goquery package](https://github.com/PuerkitoBio/goquery).

It and LinkUp can be installed with the following commands:

```shell
$ go get github.com/PuerkitoBio/goquery
$ go get github.com/hgs3/linkup
```

## License

GNU General Public License version 3. See [LICENSE](LICENSE) for details.
