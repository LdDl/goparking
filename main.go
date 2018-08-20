package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"parkingDetection/framedata"

	"parkingDetection/parklot"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"parkingDetection/pkgs/fifo"
	"parkingDetection/pkgs/inits"

	"gocv.io/x/gocv"
)

var (
	// Atomic variables for grabbing and processing functions
	boolGrab atomic.Value
	boolProc atomic.Value

	// FIFO Queue
	fifoFrames *fifo.FIFOQueue
	fifoGPS    *fifo.FIFOQueue
	wg         sync.WaitGroup

	initialParameters inits.InitParams
	boolFirstFrame    = true

	counter = 0
)

// Grabber - считывание кадров
func Grabber() {
	boolProc.Store(true)

	var webcam *gocv.VideoCapture
	var err error
	switch initialParameters.PJSON.VideoType {
	case "url":
		webcam, err = gocv.VideoCaptureFile(initialParameters.PJSON.VideoSource)
		if err != nil {
			log.Printf("Error opening video capture url: %v\n", initialParameters.PJSON.VideoSource)
			log.Fatalln(err)
			return
		}
		break
	case "device":
		deviceIndex, err := strconv.Atoi(initialParameters.PJSON.VideoSource)
		if err != nil {
			log.Fatalln(err)
			return
		}
		webcam, err = gocv.VideoCaptureDevice(deviceIndex)
		if err != nil {
			log.Printf("Error opening video device index: %v\n", deviceIndex)
			log.Fatalln(err)
			return
		}
		break
	}

	defer webcam.Close()

	var singleFrame framedata.FrameData
	singleFrame.FrameMatTrue = gocv.NewMat()
	defer singleFrame.FrameMatTrue.Close()
	singleFrame.FrameMat = gocv.NewMat()
	defer singleFrame.FrameMat.Close()
	singleFrame.FrameCounter = 0

	// Read fisrt frame and recalculate parameters depending on image scale
	if ok := webcam.Read(&singleFrame.FrameMatTrue); !ok {
		// log.Printf("Error cannot read URL or device: %v\n", initialParameters.PJSON.VideoSource)
		log.Fatalf("Error cannot read URL or device: %v\n", initialParameters.PJSON.VideoSource)
		return
	}
	singleFrame.FrameMatScale = singleFrame.FrameMatTrue.Cols() / initialParameters.PJSON.ImageResizing[0]

	for i := range initialParameters.ParkingLots {
		var tmp parklot.Lot
		tmp.SetID(initialParameters.ParkingLots[i].GetID())
		points := initialParameters.ParkingLots[i].GetContourPoints()[0]
		for j := range points {
			points[j].X /= singleFrame.FrameMatScale
			points[j].Y /= singleFrame.FrameMatScale
		}
		tmp.SetContourPoints(points)
		tmp.CalcBoundingRect()
		initialParameters.ParkingLots[i] = tmp
	}

	for boolProc.Load() == true {
		singleFrame.FrameCounter++
		if ok := webcam.Read(&singleFrame.FrameMatTrue); !ok {
			// log.Printf("Error cannot read URL or device: %v\n", initialParameters.PJSON.VideoSource)
			log.Fatalf("Error cannot read URL or device: %v\n", initialParameters.PJSON.VideoSource)
			return
		}
		if singleFrame.FrameMatTrue.Empty() {
			continue
		}

		// Resize input image
		gocv.Resize(singleFrame.FrameMatTrue, &singleFrame.FrameMat, image.Point{initialParameters.PJSON.ImageResizing[0], initialParameters.PJSON.ImageResizing[1]}, 0.0, 0.0, gocv.InterpolationDefault)
		singleFrame.Buf, err = gocv.IMEncode(".bmp", singleFrame.FrameMat)
		if err != nil {
			// Do not handle error
		}
		// Push image to queue
		fifoFrames.Push(singleFrame)
		time.Sleep(10 * time.Millisecond)
	}
	boolProc.Store(false)

	wg.Done()
}

// Processing - отправка на обработку кадров видеопотока
func Processing() {

	// Defer exits
	var singleFrame framedata.FrameData

	for boolProc.Load() == true {
		var interFrame interface{}
		var ok bool

		// Pop input image from queue
		interFrame = fifoFrames.Pop()
		if _, ok = interFrame.(framedata.FrameData); ok {
			singleFrame = interFrame.(framedata.FrameData)
		} else {
			continue
		}
		ProcessingData(&singleFrame)
	}
	wg.Done()
}

// ProcessingData - обработка кадров видеопотока
func ProcessingData(f *framedata.FrameData) {

	var imgCopy framedata.FrameData
	imgCopy = (*f).Clone()
	defer imgCopy.FrameMat.Close()
	defer imgCopy.FrameMatTrue.Close()

	var imgGray, imgBlur, roi, laplacian, delta gocv.Mat
	imgGray = gocv.NewMat()
	imgBlur = gocv.NewMat()
	roi = gocv.NewMat()
	laplacian = gocv.NewMat()
	delta = gocv.NewMatFromScalar(gocv.Scalar{Val1: 0, Val2: 0, Val3: 0, Val4: 0}, gocv.MatTypeCV64F)
	defer imgGray.Close()
	defer imgBlur.Close()
	defer roi.Close()
	defer laplacian.Close()
	defer delta.Close()

	gocv.CvtColor(imgCopy.FrameMat, &imgGray, gocv.ColorBGRToGray)
	gocv.GaussianBlur(imgGray, &imgBlur, image.Point{5, 5}, 3, 3, gocv.BorderDefault)

	for i := range initialParameters.ParkingLots {

		roi = imgBlur.Region(initialParameters.ParkingLots[i].GetBoundingRect())
		gocv.Laplacian(roi, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)

		laplAbs := gocv.NewMat()
		defer laplAbs.Close()

		emptyScalar := gocv.NewMatFromScalar(gocv.Scalar{Val1: 0, Val2: 0, Val3: 0, Val4: 0}, gocv.MatTypeCV64F)
		defer emptyScalar.Close()

		gocv.AbsDiff(laplacian, emptyScalar, &laplAbs) // alternative to cv::Abs(Mat)
		mask := initialParameters.ParkingLots[i].GetMask().Clone()

		/*START Alternative to  cv::mean(laplAbs, mask); START*/
		nonZeroesMask := float64(gocv.CountNonZero(mask))
		meanMask := 0.0
		for h := 0; h < mask.Rows(); h++ {
			for g := 0; g < mask.Cols(); g++ {
				if mask.GetUCharAt(h, g) > 0 {
					meanMask += laplAbs.GetDoubleAt(h, g)
				}
			}
		}
		meanMask /= nonZeroesMask
		/*END Alternative to  cv::mean(laplAbs, mask); END*/

		if meanMask > initialParameters.PJSON.Laplacian {
			log.Printf("Parking lot with id: %v is occupied", initialParameters.ParkingLots[i].GetID())
			gocv.DrawContours(&imgCopy.FrameMat, initialParameters.ParkingLots[i].GetContourPoints(), -1, color.RGBA{255, 0, 0, 0}, 1)
		} else {
			gocv.DrawContours(&imgCopy.FrameMat, initialParameters.ParkingLots[i].GetContourPoints(), -1, color.RGBA{0, 255, 0, 0}, 1)
		}

	}

	boolFirstFrame = false

	if initialParameters.ShowIm {
		initialParameters.GlobaWindow.IMShow(imgCopy.FrameMat)
		if initialParameters.GlobaWindow.WaitKey(1) >= 0 {
			return
		}
	}
}

func main() {
	log.Println("Starting program...")
	var err error
	cfgName := flag.String("cfg", "go_ip.json", "Config file path")
	flag.Parse()
	err = initialParameters.SetParams(*cfgName)
	if err != nil {
		log.Println(err)
		return
	}
	if initialParameters.ShowIm {
		initialParameters.GlobaWindow = gocv.NewWindow("Input Video")
		initialParameters.GlobaWindow.ResizeWindow(initialParameters.PJSON.ImageResizing[0], initialParameters.PJSON.ImageResizing[1])
		defer initialParameters.GlobaWindow.Close()
	}

	fifoFrames = fifo.NewQueue(60)
	fifoGPS = fifo.NewQueue(60)

	boolGrab.Store(true)
	boolProc.Store(true)

	wg.Add(1)
	go Grabber()

	wg.Add(1)
	go Processing()

	wg.Wait()

	boolGrab.Store(false)
	boolProc.Store(false)

	log.Println("Done!")
}
