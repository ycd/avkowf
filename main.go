package main

import (
	"bytes"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/szeliga/goray/engine"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var CommonHomepages = []string{"/", "default.aspx", "index.html", "home.html"}

func ScrapeContent(website string) string {
	hasPrefixHTTP := strings.HasPrefix(website, "http")
	if !hasPrefixHTTP {
		website = "https://" + website
	}

	resp, err := soup.Get(website)

	if err != nil {
		os.Exit(1)
	}

	*&CommonHomepages = append(*&CommonHomepages, website)

	return resp
}

func ParseHTML(unparsedHTML, domain string) string {
	doc := soup.HTMLParse(unparsedHTML)

	links := doc.FindAll("a")

	for _, link := range links {
		for _, innerElement := range link.Attrs() {
			for _, common := range CommonHomepages {
				contains := strings.Contains(common, innerElement)
				if contains {
					if innerElement != "" {
						sources := link.Find("img").HTML()
						extracted := ExtractSrcFromString(sources)
						return CheckURLContainsDomain(extracted, domain)
					}
				}
			}
		}
	}
	return "Could not find the logo link"
}

func CleanDomainName(actualLink string) string {
	u, err := url.Parse(actualLink)
	if err != nil {
		panic(err)
	}

	return u.Host
}

func ExtractSrcFromString(toExtract string) string {
	re := regexp.MustCompile("src\\s*=\\s*\"([^\"]+)\"") // Matches src="*"
	match := re.FindStringSubmatch(toExtract)

	return CleanExtractedString(match[0])
}

func CleanExtractedString(extracted string) string {
	replacer := strings.NewReplacer("src=\"", "", "\"", "")
	cleaned := replacer.Replace(extracted)
	return cleaned
}

func CheckURLContainsDomain(logoLink string, domain string) string {
	endsWithSlash := strings.HasSuffix(logoLink, "/")
	startsWithHTTP := strings.HasPrefix(logoLink, "http")
	domain = "https://" + CleanDomainName(domain) // Use cleaned domain

	if startsWithHTTP {
		return logoLink
	}
	if endsWithSlash {
		return domain + logoLink
	}

	logolink := "/" + logoLink

	return domain + logolink

}

func DownloadFileAndSave(filepath string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()


	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()


	_, err = io.Copy(out, resp.Body)
	return err
}

func DownloadFileInMemory(url string) ([]byte, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	return data, err
}

func AverageImageColor(i image.Image) color.Color {
	var r, g, b uint32

	bounds := i.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pr, pg, pb, _ := i.At(x, y).RGBA()

			r += pr
			g += pg
			b += pb
		}
	}

	d := uint32(bounds.Dy() * bounds.Dx())

	r /= d
	g /= d
	b /= d

	return color.RGBA{uint8(g), uint8(b), uint8(r), 200}
}


func CreateRandomImage(c color.Color) {
	width := 200
	height := 200
	scene := engine.NewScene(width, height)

	scene.EachPixel(func(x, y int) color.RGBA {
		return c.(color.RGBA)
	})
	scene.Save(fmt.Sprintf("%d.png", time.Now().Unix()))
}

func main() {
	link := os.Args[1]
	
	resp := ScrapeContent(link)
	logoLink := ParseHTML(resp, link)

	file, err := DownloadFileInMemory(logoLink)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		panic(err)
	}

	avgColor := AverageImageColor(img)

	CreateRandomImage(avgColor)
}
