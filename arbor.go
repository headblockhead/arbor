package arbor

import (
	"embed"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/fogleman/gg"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/golang/freetype/truetype"
	"github.com/oliamb/cutter"
	"golang.org/x/image/font"
)

//go:embed fonts
var fonts embed.FS

type Data struct {
	Name         string
	ProfileImage image.Image
	Points       string
	Attendance   string
	Week         string
}

type Creds struct {
	URL            string `json:"url"`
	User           string `json:"user"`
	Pass           string `json:"pass"`
	DeviceID       string `json:"device_id"`
	DevicePassword string `json:"device_password"`
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

func GetArborData(c *Creds) (data Data, err error) {
	var d Data

	fmt.Println("Finding browser...")
	path, _ := launcher.LookPath()
	fmt.Println("Browser found:", path)

	fmt.Println("Launching browser...")
	u := launcher.New().Bin(path).MustLaunch()
	fmt.Println("Browser launched!")
	fmt.Println("Connecting to browser...")
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	fmt.Println("Opening page...")
	page := browser.MustPage(c.URL)
	fmt.Println("Page opened!")

	// Login
	fmt.Println("Logging in...")
	fmt.Println("Entering username...")
	page.MustElement("#username").MustInput(c.User)
	fmt.Println("Entering password...")
	page.MustElement("#password").MustInput(c.Pass)
	fmt.Println("Clicking login button...")
	page.MustElement(".login-submit-button").MustClick()
	fmt.Println("Logged in!")

	// Get the Name
	fmt.Println("Getting name...")
	d.Name = page.MustElement(".sign-out-text").MustText()
	fmt.Println("Name got:", d.Name)

	// Get the profile image
	fmt.Println("Getting profile image...")
	imgStr := page.MustElement(".mis-info-panel-picture-inner").MustEval(`() => this.style.backgroundImage`).String()
	fmt.Println("Profile image src got:", imgStr)
	fmt.Println("Filtering image src...")
	r1 := regexp.MustCompile("^url[\\(]\"|\\)$|width\\/.*$") // This is a regex to remove the url() and width part of the string
	imgStr = r1.ReplaceAllString(imgStr, "")
	fmt.Println("Image src filtered:", imgStr)
	fmt.Println("Downloading image...")
	resp, err := http.Get(imgStr)
	if err != nil {
		return Data{}, err
	}
	defer resp.Body.Close()
	fmt.Println("Image downloaded!")
	imgBytes := []byte{}
	_, err = resp.Body.Read(imgBytes)
	if err != nil {
		return Data{}, err
	}
	fmt.Println("Decoding image...")
	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return Data{}, err
	}
	fmt.Println("Image decoded!")
	d.ProfileImage = img
	fmt.Println("Profile Image URL:", imgStr)

	// Get the attendance
	fmt.Println("Getting attendance...")
	attendance := page.MustElement(".mis-htmlpanel-measure-value")
	fmt.Println("Found attendance element:", attendance)
	d.Attendance = attendance.MustText()
	fmt.Println("Attendance:", attendance.MustText())
	fmt.Println("Removing attendance element...")
	attendance.MustRemove()
	fmt.Println("Attendance element removed!")

	// Get the points
	fmt.Println("Getting points...")
	points := page.MustElement(".mis-htmlpanel-measure-value")
	fmt.Println("Found points element:", points)
	d.Points = points.MustText()
	fmt.Println("Points:", points.MustText())
	fmt.Println("Removing points element...")
	points.MustRemove()
	fmt.Println("Points element removed!")

	// Get the timetable / week
	fmt.Println("Getting timetable...")
	button := page.MustElement("#myitems_0")
	fmt.Println("Found MyItems button:", button)
	button.MustClick()
	fmt.Println("Clicked MyItems button!")
	button2 := page.MustElement("#mycalendar_2")
	fmt.Println("Found MyCalendar button:", button2)
	button2.MustClick()
	fmt.Println("Clicked MyCalendar button!")
	fmt.Println("Waiting for timetable to load...")
	page.MustWaitElementsMoreThan("td", 5)
	fmt.Println("Timetable loaded!")
	return d, nil
}

func GetArborImage(d *Data) (outimg image.Image, err error) {
	// Create a new blank image
	dc := gg.NewContext(400, 300)
	// Fill the background with white
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	// Add the name
	dc.SetRGB(0, 0, 0)
	fontfile, err := fonts.ReadFile("fonts/ubmr.ttf")
	if err != nil {
		return nil, err
	}
	titleFace, err := loadFontFaceReader(fontfile, 40)
	if err != nil {
		return nil, err
	}
	subtitleFace, err := loadFontFaceReader(fontfile, 20)
	if err != nil {
		return nil, err
	}
	// eventsFace, err := loadFontFaceReader(fontfile, 12)
	// if err != nil {
	// 	return nil, err
	// }
	dc.SetFontFace(titleFace)
	// Draw the word "timetable" at the top left of the image
	dc.DrawStringAnchored("Timetable", 0, 0, 0, 0.8)
	dc.SetFontFace(subtitleFace)
	dc.DrawStringAnchored("Owner: "+d.Name, 0, 55, 0, 0)
	dc.DrawStringAnchored("Attendance: "+d.Attendance, 0, 75, 0, 0)
	dc.DrawStringAnchored("Points: "+d.Points, 0, 95, 0, 0)
	dc.DrawStringAnchored("Date: "+time.Now().Format("Monday")+" - Week "+d.Week, 0, 115, 0, 0)
	dc.DrawStringAnchored("      "+time.Now().Format("02/01/2006"), 0, 135, 0, 0)
	// Draw the profile image in the top right
	// img = img.convert("1").crop((30, 30, 160, 190))
	croppedImg, err := cutter.Crop(d.ProfileImage, cutter.Config{
		Width:  130,
		Height: 160,
		Anchor: image.Point{30, 30},
		Mode:   cutter.Centered,
	})
	// Save the image to a file
	file, err := os.Create("out.jpg")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = jpeg.Encode(file, croppedImg, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}
	// Draw the image
	dc.DrawImage(croppedImg, 270, 0)

	if err != nil {
		return nil, err
	}
	dc.DrawImage(croppedImg, 270, 0)
	dc.Clip()
	return dc.Image(), nil
}
