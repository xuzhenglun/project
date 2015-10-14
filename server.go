package main

import (
	"encoding/json"
	"fmt"
	"github.com/xuzhenglun/project/GPS_Recv"
	"github.com/xuzhenglun/project/GpsHandle"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
	LatitdeH := int(data.Latitde / 100)
	LatitdeM := int(data.Latitde - float64(LatitdeH)*100)
	LatitdeS := (data.Latitde - float64(100*LatitdeH+LatitdeM)) * 60

	LongitudeH := int(data.Longitude / 100)
	LongitudeM := int(data.Longitude - float64(LongitudeH)*100)
	LongitudeS := (data.Longitude - float64(100*LongitudeH+LongitudeM)) * 60

	Latitde := float64(LatitdeH) + float64(LatitdeM)/60 + LatitdeS/3600
	Longitude := float64(LongitudeH) + float64(LongitudeM)/60 + LongitudeS/3600
	x, y, err := GpsToMars(Longitude, Latitde)
	if err != 0 {
		fmt.Printf("Error %d: Fail to Go to Mars.\n", err)
	}
	template := Template
	template = strings.Replace(template, "MAGIC_X", strconv.FormatFloat(x, 'g', 20, 64), -1)
	template = strings.Replace(template, "MAGIC_Y", strconv.FormatFloat(y, 'g', 20, 64), -1)
	fmt.Fprintf(w, "%s", template)
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
	Status int
	Result []Gps
}

func GpsToMars(x, y float64) (float64, float64, int) {
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
		} else {
			fmt.Println(err)
		}
	}
	return mars.Result[0].X, mars.Result[0].Y, mars.Status
}

const Template string = `
<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
	<meta name="viewport" content="initial-scale=1.0, user-scalable=no" />
	<style type="text/css">
		body, html,#allmap {width: 100%;height: 100%;overflow: hidden;margin:0;font-family:"微软雅黑";}
	</style>
	<script type="text/javascript" src="http://api.map.baidu.com/api?v=2.0&ak=F37d12e0fc7f53d91bbe11819a0b8626"></script>
	<title>地址解析</title>
</head>
<body>
	<div id="allmap"></div>
</body>
</html>
<script type="text/javascript">
	var map = new BMap.Map("allmap");
	var point = new BMap.Point(MAGIC_X,MAGIC_Y);
	map.centerAndZoom(point,16);
	var myGeo = new BMap.Geocoder();
	map.addOverlay(new BMap.Marker(point));
	var mapType1 = new BMap.MapTypeControl({mapTypes: [BMAP_NORMAL_MAP,BMAP_HYBRID_MAP]});
	var mapType2 = new BMap.MapTypeControl({anchor: BMAP_ANCHOR_TOP_LEFT});
	var overView = new BMap.OverviewMapControl();
	var overViewOpen = new BMap.OverviewMapControl({isOpen:true, anchor: BMAP_ANCHOR_BOTTOM_RIGHT});
	//添加地图类型和缩略图
	map.addControl(mapType1);          //2D图，卫星图
	map.addControl(mapType2);          //左上角，默认地图控件
	map.setCurrentCity("北京");        //由于有3D图，需要设置城市哦
	map.addControl(overView);          //添加默认缩略地图控件
	map.addControl(overViewOpen);      //右下角，打开

</script>
`
