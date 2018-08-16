package inits

import (
	"encoding/json"
	"errors"
	"image"
	"io/ioutil"
	"parkingDetection/parklot"

	"gocv.io/x/gocv"
)

// InitParams - Initial parameters for programm
type InitParams struct {
	// OutputImage
	GlobaWindow *gocv.Window
	ShowIm      bool
	Areas       [][]image.Point
	ParkingLots []parklot.Lot
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
		for _, pnt := range (*ip).PJSON.Areas[i].Coords {
			localpoints = append(localpoints, image.Point{X: pnt[0], Y: pnt[1]})
		}
		(*ip).Areas = append((*ip).Areas, localpoints)
		var points []image.Point
		points = append(points, (localpoints[0]))
		points = append(points, (localpoints[1]))
		points = append(points, (localpoints[2]))
		points = append(points, (localpoints[3]))
		var tmp parklot.Lot
		tmp.ID = (*ip).PJSON.Areas[i].ID
		tmp.SetPoints(points)
		tmp.CalcBoundingRect()
		(*ip).ParkingLots = append((*ip).ParkingLots, tmp)
	}

	return err
}

// paramsJSON - sturct for parsing configuration file
type paramsJSON struct {
	VideoType     string  `json:"videoType"`
	VideoSource   string  `json:"videoSource"`
	ImageResizing []int   `json:"imageResizing"`
	ShowImage     bool    `json:"showImage"`
	Laplacian     float64 `json:"laplacian"`
	Areas         []struct {
		ID     int     `json:"id"`
		Coords [][]int `json:"coords"`
	} `json:"areas"`
}
