package globals

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type Node struct {
	Name    string
	Type    string
	Comment string
}

type SymbolConfiguration struct {
	Header struct {
		ProjectInfo struct {
			Appname    string `xml:"appname,attr"`
			Devicename string `xml:"devicename,attr"`
			Name       string `xml:"name,attr"`
		} `xml:"ProjectInfo"`
		SymbolConfigObject struct {
			Compiler   string `xml:"compiler,attr"`
			Libversion string `xml:"libversion,attr"`
			Lmm        string `xml:"lmm,attr"`
			Profile    string `xml:"profile,attr"`
			Runtimeid  string `xml:"runtimeid,attr"`
			Settings   string `xml:"settings,attr"`
			Version    string `xml:"version,attr"`
		} `xml:"SymbolConfigObject"`
		Version string `xml:"Version"`
	} `xml:"Header"`
	NodeList struct {
		Node struct {
			Name      string `xml:"name,attr"`
			Nodeflags string `xml:"nodeflags,attr"`
			Node      []struct {
				Name      string `xml:"name,attr"`
				Comment   string `xml:"Comment,omitempty"`
				Nodeflags string `xml:"nodeflags,attr"`
				Attribute string `xml:"Attribute,attr"`
				Node      []struct {
					Access    string `xml:"access,attr"`
					Name      string `xml:"name,attr"`
					Nodeflags string `xml:"nodeflags,attr"`
					Type      string `xml:"type,attr"`
					Comment   string `xml:"Comment,omitempty"`
				} `xml:"Node"`
			} `xml:"Node"`
		} `xml:"Node"`
	} `xml:"NodeList"`
	TypeList struct {
		TypeArray []struct {
			Basetype   string `xml:"basetype,attr"`
			Iecname    string `xml:"iecname,attr"`
			Name       string `xml:"name,attr"`
			Nativesize int    `xml:"nativesize,attr"`
			Size       int    `xml:"size,attr"`
			Typeclass  string `xml:"typeclass,attr"`
			ArrayDim   []struct {
				Maxrange int `xml:"maxrange,attr"`
				Minrange int `xml:"minrange,attr"`
			} `xml:"ArrayDim"`
		} `xml:"TypeArray"`
		TypeSimple []struct {
			Iecname   string `xml:"iecname,attr"`
			Name      string `xml:"name,attr"`
			Size      int    `xml:"size,attr"`
			Swapsize  int    `xml:"swapsize,attr"`
			Typeclass string `xml:"typeclass,attr"`
		} `xml:"TypeSimple"`
		TypeUserDef []struct {
			Basetype       *string  `xml:"basetype,attr"`
			Iecname        string   `xml:"iecname,attr"`
			Name           string   `xml:"name,attr"`
			Nativesize     int      `xml:"nativesize,attr"`
			Pouclass       *string  `xml:"pouclass,attr"`
			Size           int      `xml:"size,attr"`
			Swapsize       *int     `xml:"swapsize,attr"`
			Typeclass      string   `xml:"typeclass,attr"`
			Attribute      []string `xml:"Attribute"`
			Comment        *string  `xml:"Comment"`
			UserDefElement []struct {
				Byteoffset *int    `xml:"byteoffset,attr"`
				Enumvalue  *int    `xml:"enumvalue,attr"`
				Iecname    string  `xml:"iecname,attr"`
				Type       string  `xml:"type,attr"`
				Vartype    *string `xml:"vartype,attr"`
				Comment    *string `xml:"Comment"`
			} `xml:"UserDefElement"`
		} `xml:"TypeUserDef"`
	} `xml:"TypeList"`
}

func (symbole *SymbolConfiguration) parse(symbolConfigFileName string) (err error) {
	// XML-Datei öffnen
	fx, err := os.Open(symbolConfigFileName)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer fx.Close()

	// XML-Datei lesen
	byteValue, err := io.ReadAll(fx)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if err := xml.Unmarshal(byteValue, symbole); err != nil {
		return fmt.Errorf("error unmarshalling XML: %w", err)
	}

	return nil
}

func (symbole *SymbolConfiguration) GetHeaderInfo() string {
	return fmt.Sprintf("\n🪑 Name: %s\n🪑 Compiler: %s\n🪑 Profile: %s\n",
		symbole.Header.ProjectInfo.Name,
		symbole.Header.SymbolConfigObject.Compiler,
		symbole.Header.SymbolConfigObject.Profile)
}
