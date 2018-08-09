package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"parkingDetection/blobie"
	"parkingDetection/framedata"
	"parkingDetection/gpsdata"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"fifo/fifobuffer_v2"
	"parkingDetection/pkgs/inits"

	gpsd "github.com/stratoberry/go-gpsd"
	"gocv.io/x/gocv"
)

var (
	// Atomic variables for grabbing and processing functions
	boolGrab    atomic.Value
	boolProc    atomic.Value
	boolProcGPS atomic.Value

	// FIFO Queue
	fifoFrames *fifobuffer_v2.FIFOQueue
	fifoGPS    *fifobuffer_v2.FIFOQueue
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
		// Saved true image before resize
		singleFrame.FrameMat = singleFrame.FrameMatTrue.Clone()
		singleFrame.FrameMatScale = singleFrame.FrameMatTrue.Cols() / initialParameters.PJSON.ImageResizing[0]
		// Resize input image
		gocv.Resize(singleFrame.FrameMat, &singleFrame.FrameMat, image.Point{initialParameters.PJSON.ImageResizing[0], initialParameters.PJSON.ImageResizing[1]}, 0.0, 0.0, gocv.InterpolationDefault)
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

// ProcessingGPS - отправка на обработку данных с USB bu-353 GPS (Glonass)
func ProcessingGPS() {
	var gps *gpsd.Session
	var err error
	if gps, err = gpsd.Dial(gpsd.DefaultAddress); err != nil {
		log.Printf("Failed to connect to GPSD: %s", err)
		// /return
	}
	log.Printf("Succesfully to connect to GPSD: %v", gps)
	var singleGPS gpsdata.GPSData
	gps.AddFilter("TPV", func(r interface{}) {
		tpv := r.(*gpsd.TPVReport)
		singleGPS.GPSTime = tpv.Time.Add(3 * time.Hour)
		singleGPS.GPSLatitude = tpv.Lat
		singleGPS.GPSLongitude = tpv.Lon
		fifoGPS.Push(singleGPS)
	})
	done := gps.Watch()
	<-done
	wg.Done()
}

// Processing - отправка на обработку кадров видеопотока
func Processing() {

	// Defer exits
	var singleFrame framedata.FrameData
	var singleGPS gpsdata.GPSData
	var allBlobies blobie.Blobies

	for boolProc.Load() == true {
		var interFrame interface{}
		var interGPS interface{}
		var ok bool

		// Pop input image from queue
		interFrame = fifoFrames.Pop()
		if _, ok = interFrame.(framedata.FrameData); ok {
			singleFrame = interFrame.(framedata.FrameData)
		} else {
			continue
		}

		// Pop GPS from queue without checking
		interGPS = fifoGPS.Pop()
		if _, ok = interGPS.(gpsdata.GPSData); ok {
			singleGPS = interGPS.(gpsdata.GPSData)
		} else {
			// continue
		}
		ProcessingData(&singleFrame, &singleGPS, &allBlobies)
	}
	wg.Done()
}

// ProcessingData - обработка кадров видеопотока
func ProcessingData(f *framedata.FrameData, gps *gpsdata.GPSData, allBlobies *blobie.Blobies) {

	var imgCopy framedata.FrameData
	imgCopy = (*f).Clone()
	defer imgCopy.FrameMat.Close()
	defer imgCopy.FrameMatTrue.Close()

	if boolFirstFrame {
		for i := range initialParameters.Areas {
			for j := range initialParameters.Areas[i] {
				initialParameters.Areas[i][j].X = initialParameters.Areas[i][j].X / imgCopy.FrameMatScale
				initialParameters.Areas[i][j].Y = initialParameters.Areas[i][j].Y / imgCopy.FrameMatScale
			}
		}
	}

	gocv.DrawContours(&imgCopy.FrameMat, initialParameters.Areas, -1, color.RGBA{0, 255, 0, 0}, 1)
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

	fifoFrames = fifobuffer_v2.NewQueue(60)
	fifoGPS = fifobuffer_v2.NewQueue(60)

	boolGrab.Store(true)
	boolProc.Store(true)
	boolProcGPS.Store(true)

	wg.Add(1)
	go Grabber()

	wg.Add(1)
	go Processing()

	wg.Add(1)
	go ProcessingGPS()

	wg.Wait()

	boolGrab.Store(false)
	boolProc.Store(false)
	boolProcGPS.Store(false)

	log.Println("Done!")
}
