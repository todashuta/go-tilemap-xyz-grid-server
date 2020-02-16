package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/math/fixed"
)

var pathReg = regexp.MustCompile(`^/(\d+)/(\d+)/(\d+)\.png$`)

func myLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%q %q %q\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func GenerateTileImg(z, x, y int) (b []byte, err error) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	rect := img.Rect
	cBlack := color.Black
	cWhite := color.White
	for h := 0; h < rect.Max.X; h++ {
		img.Set(h, 1, cWhite)
		img.Set(h, rect.Max.Y-2, cWhite)
	}
	for v := 0; v < rect.Max.Y; v++ {
		img.Set(1, v, cWhite)
		img.Set(rect.Max.X-2, v, cWhite)
	}
	for h := 0; h < rect.Max.X; h++ {
		img.Set(h, 0, cBlack)
		img.Set(h, rect.Max.Y-1, cBlack)
	}
	for v := 0; v < rect.Max.Y; v++ {
		img.Set(0, v, cBlack)
		img.Set(rect.Max.X-1, v, cBlack)
	}

	ft, err := truetype.Parse(gomonobold.TTF)
	if err != nil {
		return b, err
	}

	opt := truetype.Options{
		Size: 20,
	}
	face := truetype.NewFace(ft, &opt)
	text := fmt.Sprintf("/%v/%v/%v", z, x, y)

	offset := fixed.I(1)

	// Background
	dr1 := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(cWhite),
		Face: face,
		Dot:  fixed.Point26_6{},
	}
	dr1.Dot.X = ((fixed.I(256) - dr1.MeasureString(text)) / 2) - offset
	dr1.Dot.Y = fixed.I(35) - offset
	dr1.DrawString(text)

	dr2 := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(cWhite),
		Face: face,
		Dot:  fixed.Point26_6{},
	}
	dr2.Dot.X = ((fixed.I(256) - dr2.MeasureString(text)) / 2) + offset
	dr2.Dot.Y = fixed.I(35) - offset
	dr2.DrawString(text)

	dr3 := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(cWhite),
		Face: face,
		Dot:  fixed.Point26_6{},
	}
	dr3.Dot.X = ((fixed.I(256) - dr3.MeasureString(text)) / 2) - offset
	dr3.Dot.Y = fixed.I(35) + offset
	dr3.DrawString(text)

	dr4 := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(cWhite),
		Face: face,
		Dot:  fixed.Point26_6{},
	}
	dr4.Dot.X = ((fixed.I(256) - dr4.MeasureString(text)) / 2) + offset
	dr4.Dot.Y = fixed.I(35) + offset
	dr4.DrawString(text)

	// Foreground
	dr5 := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(cBlack),
		Face: face,
		Dot:  fixed.Point26_6{},
	}
	dr5.Dot.X = ((fixed.I(256) - dr5.MeasureString(text)) / 2)
	dr5.Dot.Y = fixed.I(35)
	dr5.DrawString(text)

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return b, err
	}

	return buffer.Bytes(), err
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if !pathReg.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 - not found")
		return
	}

	g := pathReg.FindStringSubmatch(r.URL.Path)
	z, _ := strconv.Atoi(g[1])
	x, _ := strconv.Atoi(g[2])
	y, _ := strconv.Atoi(g[3])
	//fmt.Printf("z: %+v, x: %+v, y: %+v\n", z, x, y)

	imgbytes, err := GenerateTileImg(z, x, y)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500 - internal error")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(imgbytes)))
	if _, err := w.Write(imgbytes); err != nil {
		log.Println(err)
	}
}

func main() {
	port := flag.String("p", "", "port number")
	flag.Parse()
	if *port == "" {
		*port = "5001"
	}

	http.HandleFunc("/", rootHandler)

	log.Fatal(http.ListenAndServe(":"+*port, myLog(http.DefaultServeMux)))
}
