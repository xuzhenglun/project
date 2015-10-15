package main

import (
	"encoding/json"
	"fmt"
	"github.com/xuzhenglun/project/GPS_Recv"
	"github.com/xuzhenglun/project/GpsHandle"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const HTTP_PORT string = "80"

var data GpsHandle.GPRMC

func main() {
	data = GpsHandle.GPRMC{Status: 'V'}
	var server GPS_Recv.ServerUdp
	server.Port = 8080
	server.Info = make(chan []byte, 5)
	go server.Listen()
	go httpserver()
	for {
		gpsDataHandle(<-server.Info)
	}
}

func httpserver() {
	http.HandleFunc("/", showInGoogleMap)
	http.HandleFunc("/api", returnApi)
	http.HandleFunc("/baidu", showInBaiduMap)
	err := http.ListenAndServe(":"+HTTP_PORT, Log(http.DefaultServeMux))
	if err != nil {
		log.Printf("Error :Bind in %s Port Failed \n", HTTP_PORT)
		os.Exit(0)
	}
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func showInGoogleMap(w http.ResponseWriter, r *http.Request) {
	Latitde, Longitude := data.RTD()
	log.Printf("{%v,%v}\n", Latitde, Longitude)
	http.Redirect(w, r, `https://www.google.com/maps/place/`+strconv.FormatFloat(Latitde, 'g', 20, 64)+","+strconv.FormatFloat(Longitude, 'g', 20, 64), 302)
}

func showInBaiduMap(w http.ResponseWriter, r *http.Request) {
	if data.Status == 'V' {
		fmt.Fprintf(w, "%s", "Have not recvived any GPRMC\nPlease try after boot your device!")
	} else {
		Latitde, Longitude := data.RTD()
		log.Printf("{%v,%v}\n", Latitde, Longitude)
		x, y, err := GpsToMars(Longitude, Latitde)
		if err != 0 {
			log.Printf("Error %d: Fail to Go to Mars.\n", err)
		}
		template := Template
		template = strings.Replace(template, "MAGIC_X", strconv.FormatFloat(x, 'g', 20, 64), -1)
		template = strings.Replace(template, "MAGIC_Y", strconv.FormatFloat(y, 'g', 20, 64), -1)
		fmt.Fprintf(w, "%s", template)
	}
}

func returnApi(w http.ResponseWriter, r *http.Request) {
	Latitde, Longitude := data.RTD()

	type GpsJson struct {
		Latitde   float64
		Longitude float64
	}
	gps := []GpsJson{{data.Latitde, data.Longitude}, {Latitde, Longitude}}
	gpsJson, err := json.Marshal(gps)
	if err != nil {
		log.Println("Error :Inv Json")
	}
	log.Println("Request API")
	fmt.Fprintf(w, "%s", gpsJson)
}

func gpsDataHandle(d []byte) {
	err := data.DecodeData(d)
	if err != nil {
		log.Println(err)
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
	if x == 0 && y == 0 {
		return 0, 0, -1
	} else {
		var mars GpsMars
		client := &http.Client{}
		reqest, err := http.NewRequest("GET", "http://api.map.baidu.com/geoconv/v1/?coords="+strconv.FormatFloat(x, 'g', 10, 64)+","+strconv.FormatFloat(y, 'g', 10, 64)+"&from=1&to=5&ak=F37d12e0fc7f53d91bbe11819a0b8626", nil)
		if err != nil {
			log.Println("Error :NewRequest Fail")
			return 0, 0, -1
		}
		response, err := client.Do(reqest)
		if err != nil {
			log.Println("Error :Fail to get response")
			return 0, 0, -1
		}
		if response.StatusCode == 200 {
			body, _ := ioutil.ReadAll(response.Body)
			b := []byte(body)
			err := json.Unmarshal(b, &mars)
			if err != nil {
				log.Println(err)
			}
		}
		return mars.Result[0].X, mars.Result[0].Y, mars.Status
	}
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
