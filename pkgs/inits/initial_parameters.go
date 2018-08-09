package inits

import (
	"encoding/json"
	"errors"
	"image"
	"io/ioutil"

	"gocv.io/x/gocv"
)

// InitParams - Initial parameters for programm
type InitParams struct {
	// OutputImage
	GlobaWindow *gocv.Window
	ShowIm      bool
	Areas       [][]image.Point
	// JSON Structure
	PJSON paramsJSON
}

// SetParams - set initial params
func (ip *InitParams) SetParams(s string) (err error) {
	cfg, _ := ioutil.ReadFile(s)
	err = json.Unmarshal(cfg, &(*ip).PJSON)
	if err != nil {
		return err
	}
	if (*ip).PJSON.VideoType != "url" && (*ip).PJSON.VideoType != "device" {
		err = errors.New("videoType should be \"url\" or \"device\"")
		return err
	}
	(*ip).ShowIm = (*ip).PJSON.ShowImage

	for i := range (*ip).PJSON.Areas {
		var localpoints []image.Point
		for _, pnt := range (*ip).PJSON.Areas[i] {
			localpoints = append(localpoints, image.Point{X: pnt[0], Y: pnt[1]})
		}
		(*ip).Areas = append((*ip).Areas, localpoints)
	}
	return err
}

// paramsJSON - sturct for parsing configuration file
type paramsJSON struct {
	VideoType     string    `json:"videoType"`
	VideoSource   string    `json:"videoSource"`
	ImageResizing []int     `json:"imageResizing"`
	ShowImage     bool      `json:"showImage"`
	Areas         [][][]int `json:"areas"`
}
