package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

type handler func(image.Image, string, *launchImageSpec) error

type launchImageSpec struct {
	Width   int
	Height  int
	Postfix string
	Handler handler
}

type appIconSpec struct {
	Length int
	Name   string
}

var (
	launchImageSpecifications = []launchImageSpec{
		{320, 480, "~iphone.png", PadToPhoneScaleCutHandler},
		{640, 960, "@2x~iphone.png", PadToPhoneScaleCutHandler},
		{640, 1136, "-568h@2x~iphone.png", PadToPhoneScaleCutHandler},
		{750, 1334, "-667h@2x~iphone.png", PadToPhoneScaleCutHandler},
		{1242, 2208, "-736h@3x~iphone.png", PadToPhoneScaleCutHandler},
		{768, 1024, "-Portrait~ipad.png", ScaleHandler},
		{1024, 768, "-Landscape~ipad.png", RotateScaleCutHandler},
		{1536, 2048, "-Portrait@2x~ipad.png", SkipHandler},
		{2048, 1536, "-Landscape@2x~ipad.png", RotateScaleCutHandler},
	}
	appIconSpecifications = []appIconSpec{
		{20, "Icon-Small-20.png"},
		{29, "Icon-29.png"},
		{29, "Icon-Small.png"},
		{30, "Icon-Small-30.png"},
		{40, "Icon-40.png"},
		{40, "Icon-Small-40.png"},
		{40, "Icon-Small-20@2x.png"},
		{50, "Icon-Small-50.png"},
		{57, "Icon.png"},
		{58, "Icon-29@2x.png"},
		{58, "Icon-Small@2x.png"},
		{60, "Icon-Small-30@2x.png"},
		{72, "Icon-72.png"},
		{76, "Icon-76.png"},
		{80, "Icon-40@2x.png"},
		{80, "Icon-Small-40@2x.png"},
		{87, "Icon-29@3x.png"},
		{100, "Icon-Small-50@2x.png"},
		{114, "Icon@2x.png"},
		{120, "Icon-40@3x.png"},
		{120, "Icon-60@2x.png"},
		{120, "Icon-120.png"},
		{144, "Icon-72@2x.png"},
		{152, "Icon-76@2x.png"},
		{180, "Icon-60@3x.png"},
		{512, "iTunesArtwork.png"},
		{1024, "iTunesArtwork@2x.png"},
	}
)

func ScaleHandler(m image.Image, savePath string, spec *launchImageSpec) error {
	bounds := m.Bounds()
	var im image.Image

	if bounds.Size().Y*spec.Width/bounds.Size().X < spec.Height {
		im = resize.Resize(uint(spec.Width), 0, m, resize.Bilinear)
		log.Println("resize by width")
	} else {
		im = resize.Resize(0, uint(spec.Height), m, resize.Bilinear)
		log.Println("resize by height")
	}

	if err := saveImage(&im, savePath, 1); err != nil {
		log.Println(savePath, err)
	}
	return nil
}

func RotateScaleCutHandler(m image.Image, savePath string, spec *launchImageSpec) error {
	im := resize.Resize(uint(spec.Width), 0, m, resize.Bilinear)
	log.Println("resize by width")

	// if the origin is larger than expected
	var err error
	im, err = cutter.Crop(im, cutter.Config{
		Width:  spec.Width,
		Height: spec.Height,
		Anchor: image.Point{0, (im.Bounds().Size().Y - spec.Height) / 2},
	})
	log.Printf("cropped to %d * %d\n", im.Bounds().Size().X, im.Bounds().Size().Y)

	if err != nil {
		log.Println(savePath, err)
		return err
	}
	if err := saveImage(&im, savePath, 1); err != nil {
		log.Println(savePath, err)
		return err
	}
	return nil
}

func PadToPhoneScaleCutHandler(m image.Image, savePath string, spec *launchImageSpec) error {
	im := resize.Resize(0, uint(spec.Height), m, resize.Bilinear)
	log.Println("resize by height")

	// if the origin is larger than expected
	var err error
	im, err = cutter.Crop(im, cutter.Config{
		Width:  spec.Width,
		Height: spec.Height,
		Anchor: image.Point{(im.Bounds().Size().X - spec.Width) / 2, 0},
	})
	log.Printf("cropped to %d * %d\n", im.Bounds().Size().X, im.Bounds().Size().Y)

	if err != nil {
		log.Println(savePath, err)
		return err
	}
	if err := saveImage(&im, savePath, 1); err != nil {
		log.Println(savePath, err)
		return err
	}
	return nil
}

func SkipHandler(m image.Image, savePath string, spec *launchImageSpec) error {
	if err := saveImage(&m, savePath, 1); err != nil {
		log.Println(savePath, err)
	}
	return nil
}

func GenerateLaunchImage(origin string) error {
	reader, err := os.Open(origin)
	if err != nil {
		log.Println(origin, err)
		return err
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Println(origin, err)
		return err
	}

	base := filepath.Base(origin)
	base = base[:strings.Index(base, ".")]

	for _, spec := range launchImageSpecifications {
		savePath := base + spec.Postfix

		log.Println("generating ", savePath)
		spec.Handler(m, savePath, &spec)
	}
	return nil
}

func GenerateAppIcon(origin string) error {
	reader, err := os.Open(origin)
	if err != nil {
		log.Println(origin, err)
		return err
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Println(origin, err)
		return err
	}

	for _, spec := range appIconSpecifications {
		im := resize.Resize(uint(spec.Length), uint(spec.Length), m, resize.Bilinear)
		if err := saveImage(&im, spec.Name, 1); err != nil {
			log.Println(spec.Name, err)
		}
	}
	return nil
}

func iOSScale(origin string, templateSize string) error {
	switch templateSize {
	case "1x":
		reader, err := os.Open(origin)
		if err != nil {
			log.Println(origin, err)
			return err
		}
		defer reader.Close()
		m, _, err := image.Decode(reader)
		if err != nil {
			log.Println(origin, err)
			return err
		}

		name := filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))] + "@2x" + filepath.Ext(origin)
		im := resize.Resize(uint(m.Bounds().Size().X*2), uint(m.Bounds().Size().Y*2), m, resize.Bilinear)
		if err := saveImage(&im, name, 1); err != nil {
			log.Println(name, err)
		}

		name = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))] + "@3x" + filepath.Ext(origin)
		im = resize.Resize(uint(m.Bounds().Size().X*3), uint(m.Bounds().Size().Y*3), m, resize.Bilinear)
		if err := saveImage(&im, name, 1); err != nil {
			log.Println(name, err)
		}
	case "2x":
		var one, two, three string
		if !strings.HasSuffix(filepath.Base(origin), "@2x") {
			one = origin
			two = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))] + "@2x" + filepath.Ext(origin)
			three = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))] + "@3x" + filepath.Ext(origin)
			os.Rename(origin, two)
		} else {
			one = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))-3] + filepath.Ext(origin)
			two = origin
			three = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))-3] + "@3x" + filepath.Ext(origin)
		}

		reader, err := os.Open(two)
		if err != nil {
			log.Println(two, err)
			return err
		}
		defer reader.Close()
		m, _, err := image.Decode(reader)
		if err != nil {
			log.Println(two, err)
			return err
		}

		im := resize.Resize(uint(m.Bounds().Size().X/2), uint(m.Bounds().Size().Y/2), m, resize.Bilinear)
		if err := saveImage(&im, one, 1); err != nil {
			log.Println(one, err)
		}

		im = resize.Resize(uint(m.Bounds().Size().X*3/2), uint(m.Bounds().Size().Y*3/2), m, resize.Bilinear)
		if err := saveImage(&im, three, 1); err != nil {
			log.Println(three, err)
		}
	case "3x":
		var one, two, three string
		if !strings.HasSuffix(filepath.Base(origin), "@3x") {
			one = origin
			two = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))] + "@2x" + filepath.Ext(origin)
			three = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))] + "@3x" + filepath.Ext(origin)
			os.Rename(origin, three)
		} else {
			one = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))-3] + filepath.Ext(origin)
			two = filepath.Dir(origin) + string(filepath.Separator) + filepath.Base(origin)[:len(filepath.Base(origin))-len(filepath.Ext(origin))-3] + "@2x" + filepath.Ext(origin)
			three = origin
		}

		reader, err := os.Open(three)
		if err != nil {
			log.Println(three, err)
			return err
		}
		defer reader.Close()
		m, _, err := image.Decode(reader)
		if err != nil {
			log.Println(three, err)
			return err
		}

		im := resize.Resize(uint(m.Bounds().Size().X/3), uint(m.Bounds().Size().Y/3), m, resize.Bilinear)
		if err := saveImage(&im, one, 1); err != nil {
			log.Println(one, err)
		}

		im = resize.Resize(uint(m.Bounds().Size().X*2/3), uint(m.Bounds().Size().Y*2/3), m, resize.Bilinear)
		if err := saveImage(&im, two, 1); err != nil {
			log.Println(two, err)
		}
	default:
		log.Fatal("unrecognized template size")
	}

	return nil
}
