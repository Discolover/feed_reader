package atom

import (
	"fmt"
	"encoding/xml"
	"io"
	"os"
	"time"
)

type MediaType string

type TextConstructType string

const (
	Text TextConstructType = "text"
	HTML = "html"
	XHTML = "xhtml"
)

func (t TextConstructType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var typeRaw string

	err := d.DecodeElement(&typeRaw, &start)
	if err != nil {
		panic(err)
	}

	types := map[string]struct{}{
		"text": struct{}{},
		"html": struct{}{},
		"xhtml": struct{}{},
	}

	if _, isExist := types[typeRaw]; isExist {
		t = TextConstructType(typeRaw)
	} else {
		panic(fmt.Sprintf("Unexisting type %s", typeRaw))
	}

	return err
}

type IRI string 
type EmailAddress string 

type CommonAttributes struct {
	Base IRI `xml:"xml:base,attr"`
	Lang string `xml:"xml:lang,attr"`
}

type TextConstruct struct {
	CommonAttributes
	Type TextConstructType `xml:"type"`
	Data string `"xml:,innerxml"`
}

type PersonConstruct struct {
	CommonAttributes
	Name string `xml:"name"`
	URI IRI `xml:"uri"`
	Email EmailAddress `xml:"email"`
}

type DateConstruct struct {
	CommonAttributes
	DateTime time.Time
}

func (dc *DateConstruct) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var dateTimeRaw string
	err := d.DecodeElement(&dateTimeRaw, &start)
	if err != nil {
		panic(err)
	}

	dc.DateTime, err = time.Parse(time.RFC3339, dateTimeRaw)
	if err != nil {
		panic(err)
	}

	return err
}

type Category struct {
	Term string `xml:"term,attr"`
	Scheme IRI `xml:"scheme,attr"`
	Label string `xml:"label,attr"`
}

type Generator struct {
	CommonAttributes
	URI IRI `xml:"uri,attr"`
	Version string `xml:"version,attr"`
	Text string `"xml:,innerxml"`
}

type Icon struct {
	CommonAttributes
	Value IRI `xml:",innerxml"`
}

type Logo struct {
	CommonAttributes
	Value IRI `xml:",innerxml"`
}

// rfc4287: Note that the definition of "IRI" excludes relative references.
type ID struct {
	CommonAttributes
	Value IRI `xml:",innerxml"`
}

type Link struct {
	CommonAttributes
	Href IRI `xml:"href,attr"`
	Rel string `xml:"rel,attr"` // NCName | IRI
	Type MediaType `xml:"type,attr"`
	HrefLang string `xml:"hreflang,attr"`
	Title string `xml:"title,attr"`
	Length int `xml:"length,attr"`
}

type Feed struct {
	CommonAttributes
	XMLName xml.Name `xml:"feed"`
	Author []PersonConstruct `xml:"author"`
	Category []Category `xml:"category"`
	Contributor []PersonConstruct `xml:"contributor"`
	Generator Generator `xml:"generator"`
	Icon Icon `xml:"icon"`
	ID ID `xml:"id"`
	Link []Link `xml:"link"`
	Logo Logo `xml:"logo`
	Rights TextConstruct `xml:"rights"`
	Subtitle TextConstruct `xml:"subtitle"`
	Title TextConstruct `xml:"title"`
	Updated DateConstruct `xml:"updated"`
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	CommonAttributes
	XMLName xml.Name `xml:"entry"`
	Author []PersonConstruct `xml:"author"`
	Category []Category `xml:"category"`
	// Content
	Contributor []PersonConstruct `xml:"contributor"`
	ID ID `xml:"id"`
	Link []Link `xml:"link"`
	Published DateConstruct `xml:"published"`
	Rights TextConstruct `xml:"rights"`
	// Source
	Summary TextConstruct `xml:"summary"`
	Title TextConstruct `xml:"title"`
	Updated DateConstruct `xml:"updated"`
	// Extension
}


func main() {
	f, err := os.Open("sample.xml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	
	var v Entry
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", v)
}