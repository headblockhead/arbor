package arbor

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"regexp"

	"github.com/fogleman/gg"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed fonts
var fonts embed.FS

//go:embed creds.json
var creds []byte

type Data struct {
	Name         string
	ProfileImage image.Image
	Points       string
	Attendance   string
}

type Creds struct {
	URL            string `json:"url"`
	User           string `json:"user"`
	Pass           string `json:"pass"`
	DeviceID       string `json:"device_id"`
	DevicePassword string `json:"device_password"`
}

func addLabel(img image.Image, x, y int, label string, size int) (outimage image.Image, err error) {
	var w = img.Bounds().Dx()
	var h = img.Bounds().Dy()
	dc := gg.NewContext(w, h)
	// Text color - white
	dc.SetRGB(1, 1, 1)

	// Load the font
	fontBytes, err := fonts.ReadFile("fonts/ubmr.ttf")
	if err != nil {
		return nil, err
	}
	face, err := loadFontFaceReader(fontBytes, float64(size))
	if err != nil {
		return nil, err
	}
	dc.SetFontFace(face)

	// Draw the background
	dc.DrawImage(img, 0, 0)
	// Draw text at position - anchor on the top left corner of the text
	dc.DrawStringAnchored(label, float64(x), float64(y), 0, 0)
	dc.Clip()

	outimage = dc.Image()
	return outimage, nil
}

func loadFontFaceReader(fontBytes []byte, points float64) (font.Face, error) {
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
		// Hinting: font.HintingFull,
	})
	return face, nil
}

func GetArborData() (data Data, err error) {
	var c Creds
	var d Data

	err = json.Unmarshal(creds, &c)
	if err != nil {
		return Data{}, err
	}

	path, _ := launcher.LookPath()
	u := launcher.New().Bin(path).MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(c.URL)

	// Login
	page.MustElement("#username").MustInput(c.User)
	page.MustElement("#password").MustInput(c.Pass)
	page.MustElement(".login-submit-button").MustClick()

	// Get the Name
	d.Name = page.MustElement(".sign-out-text").MustText()

	fmt.Println("Name:", d.Name)

	// Get the profile image
	imgStr := page.MustElement(".mis-info-panel-picture-inner").MustEval(`() => this.style.backgroundImage`).String()
	r1 := regexp.MustCompile("^url[\\(]\"|\\)$|width\\/.*$") // This is a regex to remove the url() and width part of the string
	imgStr = r1.ReplaceAllString(imgStr, "")
	resp, err := http.Get(imgStr)
	if err != nil {
		return Data{}, err
	}
	defer resp.Body.Close()
	imgBytes := []byte{}
	_, err = resp.Body.Read(imgBytes)
	if err != nil {
		return Data{}, err
	}
	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return Data{}, err
	}
	d.ProfileImage = img

	fmt.Println("Profile Image:", imgStr)

	// Get the attendance
	attendance := page.MustElement(".mis-htmlpanel-measure-value")
	fmt.Println("Attendance:", attendance.MustText())
	d.Attendance = attendance.MustText()
	attendance.MustRemove()

	// Get the points
	points := page.MustElement(".mis-htmlpanel-measure-value")
	fmt.Println("Points:", points.MustText())
	d.Points = points.MustText()
	points.MustRemove()

	fmt.Println(d)

	return Data{}, nil
}
