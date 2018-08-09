package framedata

import (
	"image"

	"gocv.io/x/gocv"
)

// DetectObjectsHaar - object detection via Haar cascades
func (fd *FrameData) DetectObjectsHaar(cascade *gocv.CascadeClassifier, scale float64, minNeighbors int, minSize int, maxSize int) (ret []image.Rectangle) {
	// var matGray gocv.Mat
	// matGray = (*fd).FrameMat.Clone()
	// gocv.CvtColor((*fd).FrameMat, &matGray, gocv.ColorBGRToGray)

	ret = (*cascade).DetectMultiScaleWithParams((*fd).FrameMat, scale, minNeighbors, 0, image.Point{minSize, minSize}, image.Point{maxSize, maxSize})
	// ret = (*cascade).DetectMultiScale((*fd).FrameMatTrue)
	return ret
}


