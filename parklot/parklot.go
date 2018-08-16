package parklot

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

type Lot struct {
	ID                          int
	Occupied                    bool
	ContourPoints               [][]image.Point
	PolygonPointsInBoundingRect []image.Point
	BoudingRect                 image.Rectangle
	Mask                        gocv.Mat
}

func NewParkingLot() Lot {
	return Lot{
		ID:       -1,
		Occupied: false,
	}
}

func (l *Lot) SetStatus(s bool) {
	(*l).Occupied = s
}

func (l *Lot) GetStatus() bool {
	return (*l).Occupied
}

func (l *Lot) SetID(i int) {
	(*l).ID = i
}

func (l *Lot) GetID() int {
	return (*l).ID
}

func (l *Lot) SetPoints(p []image.Point) {
	var contours [][]image.Point
	contours = append(contours, p)
	(*l).ContourPoints = contours
}

func (l *Lot) GetContourPoints() [][]image.Point {
	return (*l).ContourPoints
}

func (l *Lot) CalcBoundingRect() {
	if len(((*l).ContourPoints)) != 0 {
		(*l).BoudingRect = gocv.BoundingRect((*l).ContourPoints[0])
		(*l).Mask = gocv.NewMatWithSize((*l).BoudingRect.Size().Y, (*l).BoudingRect.Size().X, gocv.MatTypeCV8UC1)
		for _, p := range (*l).ContourPoints[0] {
			(*l).PolygonPointsInBoundingRect = append((*l).PolygonPointsInBoundingRect, image.Point{X: p.X - (*l).BoudingRect.Min.X, Y: p.Y - (*l).BoudingRect.Min.Y})
		}
		var contours [][]image.Point
		contours = append(contours, (*l).PolygonPointsInBoundingRect)
		gocv.DrawContours(&(*l).Mask, contours, -1, color.RGBA{255, 255, 255, 255}, -1)
	} else {

	}
}

func (l *Lot) GetBoundingRect() image.Rectangle {
	return (*l).BoudingRect
}

func (l *Lot) GetMask() *gocv.Mat {
	return &l.Mask
}

func (l *Lot) GetCenterPoint() image.Point {
	return image.Point{X: (2 + (*l).BoudingRect.Min.X + (*l).BoudingRect.Dx()) / 2, Y: (2 + (*l).BoudingRect.Min.Y + (*l).BoudingRect.Dy()) / 2}
}
