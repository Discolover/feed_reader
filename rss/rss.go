// TODO:
// - add other namespaces fields
// - add custom parsing mechanism for `Link`

package rss

import (
	//"fmt"
	"log"
	"encoding/xml"
	//"encoding/json"
	//"io"
	"strings"
	//"os"
	"time"
	"strconv"
)

const (
	AtomNameSpaceURL = "http://www.w3.org/2005/Atom"
)

type URL string
type Email string

type DateTime time.Time

func (dt *DateTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var timeRaw string
	err := d.DecodeElement(&timeRaw, &start)
	if err != nil {
		return err
	}

	timeRaw = strings.TrimSpace(timeRaw)

	t, err := time.Parse(time.RFC1123, timeRaw)
	if _, ok := err.(*time.ParseError); ok {
		t, err = time.Parse(time.RFC1123Z, timeRaw)
	}
	if err != nil {
		return err
	}

	*dt = DateTime(t)

	return err
}

func (dt DateTime) String() string {
	return time.Time(dt).String()
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return time.Time(dt).MarshalJSON()
}

type Image struct {
	URL URL `xml:"url"`
	Title string `xml:"title"`
	Link URL `xml:"link"`
	Width int `xml:"width"`
	Height int `xml:"height"`
	Description string `xml:"description"`
}

type Cloud struct {
	Domain string `xml:"domain,attr"` // hostname or IP
	Port int `xml:"port,attr"`
	Path string `xml:"path,attr"`
	RegisterProcedure string `xml:"registerProcedure,attr"`
	Protocol string `xml:"protocol,attr"`
}

type Source struct {
	URL URL `xml:"url,attr"`
	Value string `xml:",chardata"`
}

type Enclosure struct {
	URL URL `xml:"url,attr"`
	Length int `xml:"length,attr"`
	Type string `xml:"type,attr"` 
}

type Category struct {
	Domain string `xml:"domain,attr"`
	Value string `xml:",chardata"`
}

type Guid struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value URL `xml:",chardata"`
}

type Description struct {
	Value string `xml:",chardata"`
}

type Link struct {
	XMLName xml.Name
	// atom.Link
	Href URL `xml:"href,attr"`
	HrefLang string `xml:"hreflang,attr"`
	Length int `xml:"length,attr"`
	Title string `xml:"title,attr"`
	Type string `xml:"type,attr"`
	Rel string `xml:"rel,attr"`
	Value URL `xml:",chardata"`
}

//func (l *Link) String() string {
//	if l.XMLName.Space == "http://www.w3.org/2005/Atom" {
//		return string(l.Href)
//	}
//
//	return string(l.Value)
//}

type Item struct {
	Title string `xml:"title"`
	Links []Link `xml:"link"`
	Description Description `xml:"description"` // HTML
	Author Email `xml:"author"`
	Category []Category `xml:"category"`
	Comments URL `xml:"comments"`
	Enclosure Enclosure `xml:"enclosure"`
	Guid Guid `xml:"guid"`
	PubDate *DateTime `xml:"pubDate"`
	Source Source `xml:"source"`
}

type Channel struct {
	Title string `xml:"title"`
	// can have several `Links` due to the `atom` namespace
	Links []Link `xml:"link"`
	Description string `xml:"description"`
	Language string `xml:"language"`
	Copyright string `xml:"copyright"`
	ManagingEditor Email `xml:"managingEditor"`
	WebMaster Email `xml:"webMaster"`
	PubDate *DateTime `xml:"pubDate"`
	LastBuildDate *DateTime `xml:"lastBuildDate"`
	Categories []Category `xml:"category"`
	Generator string `xml:"generator"`
	Docs URL `xml:"docs"`
	Cloud Cloud `xml:"cloud"`
	TTL int `xml:"ttl"`
	Image Image `xml:"image"`
	Rating string `xml:"rating"`
	// `TextInput` is not implemented
	SkipHours []int `xml:"skipHours>hour"`
	SkipDays []string `xml:"skipDays>day"`

	Items []Item `xml:"item"`
}

type Document struct {
	XMLName xml.Name `xml:"rss"`
	Version string `xml:"version,attr"`
	Channel Channel `xml:"channel"`
}

func (d *Document) SelfLink() string {
	for _, l := range d.Channel.Links {
		if l.XMLName.Space == AtomNameSpaceURL && l.Rel == "self" {
			return string(l.Href)
		}
	}

	return ""
}

func (d *Document) SetSelfLink(uri string) {
	selfLink := Link{
		Type: "application/rss+xml",
		Rel: "self",
		XMLName: xml.Name{
			Local: "link",
			Space: AtomNameSpaceURL,
		},
		Href: URL(uri),
	}

	d.Channel.Links = append(d.Channel.Links, selfLink)
}

func complainIfEmpty(value, name, containerName string) {
	if strings.TrimSpace(value) == "" {
		log.Fatalf("RSS's `%s` element doesn't contain required " +
				   "`%s` element or it's empty.", containerName, name)
	}
}

func (d *Document) Postprocess() {
	channel := &d.Channel

	for i, item := range channel.Items {
		_, err := strconv.ParseBool(item.Guid.IsPermaLink)
		// if not a valid explicit bool expression, defaults to "true" 
		if err != nil {
			channel.Items[i].Guid.IsPermaLink = "true"
		}
	}
}

func (d *Document) Validate() {
	if d.Version != "2.0" {
		log.Fatalf("Not supported rss document version: \"%s\".\n" +
				   "Supports only \"2.0\".", d.Version)
	}

	channel := &d.Channel

	complainIfEmpty(channel.Description, "description", "channel")
	//ComplainIfEmpty(string(channel.Link), "link", "channel")
	complainIfEmpty(channel.Title, "title", "channel")

	//for _, item := range channel.Items {
	//	if strings.TrimSpace(item.Title) == "" &&
	//	   strings.TrimSpace(item.Description) == "" {
	//		log.Fatalf("`item` element must either contain " +
	//				   "`title` or `description` element.")
	//	}
	//}
}

func NewDocument(raw []byte) (*Document, error) {
	var d Document

	err := xml.Unmarshal(raw, &d)
	if err != nil {
		return nil, err
	}

	d.Postprocess()
	d.Validate()

	return &d, nil
}