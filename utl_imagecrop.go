package main

import (
	"fmt"
	"image"
	"io"
	"os"

	"github.com/disintegration/imaging"
)

func CropImageMan(img io.Reader, size [2]uint, BaseSavePath string) error {
	im, err := imaging.Decode(img)
	if err != nil {
		return err
	}
	//get image dimension
	dimension := im.Bounds().Size()
	SourceImageColor := im.Bounds().Bounds().At(0, 0)
	//to calucultae the times of being cutted
	var h int
	var w int
	if wtemp := dimension.X % int(size[0]); wtemp == 0 {
		w = dimension.X / int(size[0])
	} else {
		w = dimension.X/int(size[1]) + 1
	}
	if htemp := dimension.Y % int(size[1]); htemp == 0 {
		h = dimension.Y / int(size[1])
	} else {
		h = dimension.Y/int(size[1]) + 1
	}
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			TempColorImg := imaging.Crop(im, image.Rectangle{Min: image.Point{i * int(size[0]), j * int(size[1])}, Max: image.Point{(i + 1) * int(size[0]), (j + 1) * int(size[1])}})
			if ((i+1)*int(size[0]) > dimension.X) || ((j+1)*int(size[1]) > dimension.Y) {
				//create a blank image
				TempColorImgWhite := imaging.New(int(size[0]), int(size[1]), SourceImageColor)
				TempColorImg = imaging.Overlay(TempColorImgWhite, TempColorImg, image.Point{0, 0}, 1)
			}
			f, e := os.Create(fmt.Sprintf("%s-%v-%v.jpg", BaseSavePath, i, j))
			if e != nil {
				return e
			}
			imaging.Encode(f, TempColorImg, imaging.JPEG)
		}
	}
	return nil
}
