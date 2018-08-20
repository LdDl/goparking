package parklot

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

// Lot - struct for storing data for parking lots (actually it can be any other polygonal object)
type Lot struct {
	id                          string
	occupied                    bool
	contourPoints               [][]image.Point
	polygonPointsInBoundingRect []image.Point
	boudingRect                 image.Rectangle
	mask                        gocv.Mat
}

// NewParkingLot - constructor for Lot type
func NewParkingLot() Lot {
	return Lot{
		id:       "EmptyID",
		occupied: false,
	}
}

// SetStatus - sets status for lot
func (l *Lot) SetStatus(s bool) {
	(*l).occupied = s
}

// GetStatus - returns status of lot
func (l *Lot) GetStatus() bool {
	return (*l).occupied
}

// SetID - sets ID for lot
func (l *Lot) SetID(i string) {
	(*l).id = i
}

// GetID - returns ID of lot
func (l *Lot) GetID() string {
	return (*l).id
}

// SetContourPoints - sets contour points for lot
func (l *Lot) SetContourPoints(p []image.Point) {
	var contours [][]image.Point
	contours = append(contours, p)
	(*l).contourPoints = contours
}

// GetContourPoints - returns contour points of lot
func (l *Lot) GetContourPoints() [][]image.Point {
	return (*l).contourPoints
}

// GetBoundingRect - returns bouding rect of lot
func (l *Lot) GetBoundingRect() image.Rectangle {
	return (*l).boudingRect
}

// GetMask - returns mask of lot
func (l *Lot) GetMask() *gocv.Mat {
	return &l.mask
}

// GetCenterPoint - calculates and returns center point of bounding rect for lot
func (l *Lot) GetCenterPoint() image.Point {
	return image.Point{X: (2 + (*l).boudingRect.Min.X + (*l).boudingRect.Dx()) / 2, Y: (2 + (*l).boudingRect.Min.Y + (*l).boudingRect.Dy()) / 2}
}

// CalcBoundingRect - calculates and sets bounding rect for lot
func (l *Lot) CalcBoundingRect() {
	if len(((*l).contourPoints)) != 0 {
		(*l).boudingRect = gocv.BoundingRect((*l).contourPoints[0])
		(*l).mask = gocv.NewMatWithSize((*l).boudingRect.Size().Y, (*l).boudingRect.Size().X, gocv.MatTypeCV8UC1)
		for _, p := range (*l).contourPoints[0] {
			(*l).polygonPointsInBoundingRect = append((*l).polygonPointsInBoundingRect, image.Point{X: p.X - (*l).boudingRect.Min.X, Y: p.Y - (*l).boudingRect.Min.Y})
		}
		var contours [][]image.Point
		contours = append(contours, (*l).polygonPointsInBoundingRect)
		gocv.DrawContours(&(*l).mask, contours, -1, color.RGBA{255, 255, 255, 255}, -1)
	} else {

	}
}
