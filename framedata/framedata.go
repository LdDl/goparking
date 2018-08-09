package framedata

import (
	"gocv.io/x/gocv"
)

// FrameData - Struct for storing input data (image, frame counter, etc.)
type FrameData struct {
	FrameCounter  int
	FrameMat      gocv.Mat
	FrameMatTrue  gocv.Mat
	FrameMatScale int
	Buf           []byte // Encoded (mat -> bmp bytes)
}

// Clone - clone FrameData
func (fd *FrameData) Clone() FrameData {
	var newFD FrameData
	newFD.FrameCounter = (*fd).FrameCounter
	newFD.FrameMatScale = (*fd).FrameMatScale
	newFD.FrameMat = (*fd).FrameMat.Clone()
	newFD.FrameMatTrue = (*fd).FrameMatTrue.Clone()
	return newFD
}
