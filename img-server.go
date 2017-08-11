package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"

	"github.com/hoisie/web"
	"github.com/loranbriggs/go-camera"
)

var (
	c           = camera.New("/home/pi/tmp")
	currentPath = "" // current path to an image
)

func index() string {
	return `
    <html>
    <head><title>Oops!</title></head>
    <body style="text-align:center">
    <h1>Get out!</h1>
    <p>You probably meant to hit the <code>/img</code> endpoint.</p>
    </body>
    </html>
    `
}

func serveImg(ctx *web.Context) string {
	ctx.Server.Logger.Print("Serving image... ")
	defer ctx.Server.Logger.Println("Done")
	img, err := ioutil.ReadFile(currentPath)
	if err != nil {
		fmt.Println("read fail! " + err.Error())
		ctx.Abort(500, "Failed to read file: "+err.Error())
		return ""
	}
	ctx.ContentType("image/png")
	return string(img)
}

func main() {
	c.Vflip(true)
	// concurrently take pictures at an interval,
	// because the camera is slow and doing it
	// synchronously is A Bad Thing
	go func() {
		for {
			fmt.Print("Taking a picture... ")

			s, err := c.Capture()
			if err != nil {
				panic(err)
			}

			img, err := imgio.Open(s)
			if err != nil {
				fmt.Println(err)
			}

			r1 := transform.FlipV(img)
			r2 := effect.EdgeDetection(img, 1.0)
			s = s[:len(s)-5]

			if err := imgio.Save(s, r2, imgio.PNG); err != nil {
				fmt.Println(err)
			}

			if currentPath != "" {
				err := os.Remove(currentPath) // remove the previous image if there is one
				if err != nil {
					fmt.Println("failed to remove file: " + err.Error())
					panic(err)
				}
			}

			currentPath = s + ".png"
			fmt.Println("Done")

			time.Sleep(500 * time.Millisecond) // the 1 isn't strictly necessary, but it reads better this way
		}
	}()

	web.Get("/", index)
	web.Get("/img", serveImg) // `/img GET` => the image
	web.Run("0.0.0.0:8080")
}
