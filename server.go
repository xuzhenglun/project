package main

import (
	"encoding/json"
	"fmt"
	"github.com/xuzhenglun/Project/GpsHandle"
	"github.com/xuzhenglun/project/GPS_Recv"
	"io/ioutil"
	"net/http"
	"strconv"
)

var data GpsHandle.GPRMC

func main() {
	var server GPS_Recv.ServerUdp
	server.Port = 8080
	server.Info = make(chan []byte, 5)
	go server.Listen()
	go httpserver()

	for {
		item := <-server.Info
		if string(item[0:5]) == "Magic" {
			fmt.Println("Accepted")
		}
		gpsDataHandle(item)
	}
}

func httpserver() {
	http.HandleFunc("/", showInGoogleMap)
	http.HandleFunc("/api", returnApi)
	http.HandleFunc("/baidu", showInBaiduMap)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func showInGoogleMap(w http.ResponseWriter, r *http.Request) {
	LatitdeH := int(data.Latitde / 100)
	LatitdeM := int(data.Latitde - float64(LatitdeH)*100)
	LatitdeS := (data.Latitde - float64(100*LatitdeH+LatitdeM)) * 60

	LongitudeH := int(data.Longitude / 100)
	LongitudeM := int(data.Longitude - float64(LongitudeH)*100)
	LongitudeS := (data.Longitude - float64(100*LongitudeH+LongitudeM)) * 60

	Latitde := strconv.Itoa(LatitdeH) + "°" + strconv.Itoa(LatitdeM) + "'" + strconv.FormatFloat(LatitdeS, 'g', 10, 64) + `"` + string(data.SN)
	Longitude := strconv.Itoa(LongitudeH) + "°" + strconv.Itoa(LongitudeM) + "'" + strconv.FormatFloat(LongitudeS, 'g', 10, 64) + `"` + string(data.EW)

	fmt.Println(Latitde)
	fmt.Println(Longitude)

	http.Redirect(w, r, `https://www.google.com/maps/place/`+Latitde+Longitude, 302)
}

func showInBaiduMap(w http.ResponseWriter, r *http.Request) {
}

func returnApi(w http.ResponseWriter, r *http.Request) {

}

func gpsDataHandle(d []byte) {
	err := data.DecodeData(d)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(data)
	}
}

type Gps struct {
	X, Y float64
}

type GpsMars struct {
	Status bool
	Result Gps
}

func GpsToMars(x, y float64) (float64, float64, bool) {
	var mars GpsMars
	client := &http.Client{}
	reqest, _ := http.NewRequest("GET", "http://api.map.baidu.com/geoconv/v1/?coords="+strconv.FormatFloat(x, 'g', 10, 64)+","+strconv.FormatFloat(y, 'g', 10, 64)+"&from=1&to=5&ak=F37d12e0fc7f53d91bbe11819a0b8626", nil)
	response, _ := client.Do(reqest)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		b := []byte(body)
		err := json.Unmarshal(b, &mars)
		if err == nil {
			fmt.Println(mars)
		}
	}
	return mars.Result.X, mars.Result.Y, mars.Status
}
