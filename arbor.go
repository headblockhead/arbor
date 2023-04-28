package arbor

import (
	"embed"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/MaxHalford/halfgone"
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
	TimeTable    []string
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

	page.MustWaitLoad()
	page.WaitElementsMoreThan("input", 1)

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
	imgStr := page.MustElement(".info-panel-picture-inner").MustEval(`() => this.src`).String()
	fmt.Println("Profile image src got:", imgStr)
	fmt.Println("Filtering image src...")
	r1 := regexp.MustCompile("^url[\\(]\"|\\)$|width\\/.*$") // This is a regex to remove the url() and width part of the string
	imgStr = r1.ReplaceAllString(imgStr, "")
	imgStr += "width/200"
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

	fmt.Println("Finding Day button...")
	daybutton := page.MustElement(".x-segmented-button-first")
	fmt.Println("Clicking Day button...")
	daybutton.MustClick()
	fmt.Println("Clicked Day button!")

	// DEBUG: skip forward a few days
	//next := page.MustElement(".mis-calendar-navigation-button-next")
	//for i := 0; i < 6; i++ {
	//	next.MustClick()
	time.Sleep(time.Second)
	//}

	calendarContainer := page.MustElement(".mis-cal-day")
	calendar := calendarContainer.MustElement("tbody")
	calLists := calendar.MustElement("tr")
	calLists.MustElement("td").MustRemove()
	calElements := calLists.MustElement("td")
	// calElements is a list of divs that contain the date,time and name of events.

	// Loop through the elements and get the text
	fmt.Println("Getting timetable...")
	timetable := []string{}
	for _, element := range calElements.MustElements(".mis-cal-event") {
		fmt.Println("Found element:", element)
		// Revove newlines and replace with " - "
		r := regexp.MustCompile("[\\r\\n]+")
		// Remove "Location: "
		r2 := regexp.MustCompile("Location: ")
		// Replace " - " with "-"
		r3 := regexp.MustCompile(" - ")
		// Replace " | " with "|"
		r4 := regexp.MustCompile(" \\| ")
		timetable = append(timetable, r2.ReplaceAllString(r.ReplaceAllString(r4.ReplaceAllString(r3.ReplaceAllString(element.MustText(), "-"), "|"), "|"), ""))
	}
	fmt.Println("Timetable got:", timetable)
	d.TimeTable = timetable

	// Get which week it is
	fmt.Println("Getting week...")
	week := page.MustElement(".mis-calendar-title")
	fmt.Println("Found week element:", week)
	// Format: Monday 17 April 2023 (Week A)
	// Wanted: A
	weekStr := week.MustText()
	fmt.Println("Week got:", weekStr)
	fmt.Println("Filtering week...")
	r := regexp.MustCompile("Week [A-Z]")
	weekStr = r.FindString(weekStr)
	r2 := regexp.MustCompile("Week ")
	weekStr = r2.ReplaceAllString(weekStr, "")
	fmt.Println("Week filtered:", weekStr)
	d.Week = weekStr
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
	timetableFace, err := loadFontFaceReader(fontfile, 15)

	// Draw the word "timetable" at the top left of the image
	dc.SetFontFace(titleFace)
	dc.DrawStringAnchored("Timetable", 0, 0, 0, 0.8)

	dc.SetFontFace(subtitleFace)
	dc.DrawStringAnchored("Owner: "+d.Name, 0, 55, 0, 0)
	dc.DrawStringAnchored("Attendance: "+d.Attendance, 0, 75, 0, 0)
	dc.DrawStringAnchored("Points: "+d.Points, 0, 95, 0, 0)
	dc.DrawStringAnchored("Date: "+time.Now().Format("Monday")+" - Week "+d.Week, 0, 115, 0, 0)
	dc.DrawStringAnchored("      "+time.Now().Format("02/01/2006"), 0, 135, 0, 0)

	dc.SetFontFace(subtitleFace)
	dc.DrawStringAnchored("Events:", 0, 165, 0, 0)

	dc.SetFontFace(timetableFace)
	if len(d.TimeTable) == 0 {
		dc.DrawStringAnchored("No events today!", 0, 180, 0, 0)
	} else {
		for i, event := range d.TimeTable {
			dc.DrawStringAnchored(strconv.Itoa(i+1)+"|"+event, 0, float64(180+(12*i)), 0, 0)
		}
	}

	// Replace fuzzy text with crisp text
	greyText := halfgone.ImageToGray(dc.Image())
	greyText = halfgone.ThresholdDitherer{Threshold: 100}.Apply(greyText)
	dc.Clear()
	dc.DrawImage(greyText, 0, 0)

	// Crop the profile image
	croppedImg, err := cutter.Crop(d.ProfileImage, cutter.Config{
		Width:  150,
		Height: 141,
		Anchor: image.Point{101, 100},
		Mode:   cutter.Centered,
	})
	// Draw the image
	dc.DrawImage(croppedImg, 225, -30)
	if err != nil {
		return nil, err
	}
	dc.Clip()
	return dc.Image(), nil
}
