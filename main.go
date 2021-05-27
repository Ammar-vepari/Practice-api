package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

var MapData map[string]map[string]string

const radConv = math.Pi / 180.0

type cordinates struct {
	State     string  `json:"state"`
	District  string  `json:"district"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type position struct {
	Latitude  float64
	Longitude float64
}

type outputStation struct {
	Name     string
	Distance float64
}

var cord []cordinates
var PortNumber string = ":8000"

func distance(t, s position) float64 {
	dlong := (s.Longitude - t.Longitude) * radConv
	dlat := (s.Latitude - t.Latitude) * radConv
	a := math.Pow(math.Sin(dlat/2.0), float64(2)) + math.Cos(t.Latitude*radConv)*math.Cos(s.Latitude*radConv)*math.Pow(math.Sin(dlong/2.0), float64(2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := 6357 * c

	return d

}

func getCordinates(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cord)

}

func getCordinatesbyId(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	Data := mux.Vars(r)
	p := Data["data"]
	Loc := MapData[p]

	json.NewEncoder(w).Encode(Loc)

}

func getMinDistance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var pos position
	_ = json.NewDecoder(r.Body).Decode(&pos)

	target := pos

	distances := make([]outputStation, len(cord))

	for i, value := range cord {
		//fmt.Printf("target Lat:%v , Long:%v\n Source Lat:%v , Long:%v\n", pos.Latitude, pos.Longitude, value.Latitude, value.Longitude)
		distances[i] = outputStation{
			Name:     value.District,
			Distance: distance(target, position{Latitude: value.Latitude, Longitude: value.Longitude}),
		}
	}

	sort.SliceStable(distances, func(i int, j int) bool {
		return distances[i].Distance < distances[j].Distance
	})
	//fmt.Println(distances[1])
	json.NewEncoder(w).Encode(distances)

}

func getDistance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	Data := mux.Vars(r)
	p := Data["data"]
	Loc := MapData[p]
	var pos position
	_ = json.NewDecoder(r.Body).Decode(&pos)

	target := pos

	var d float64
	for k, v := range Loc {
		lat2, _ := strconv.ParseFloat(k, 64)
		long2, _ := strconv.ParseFloat(v, 64)

		source := position{Latitude: lat2, Longitude: long2}
		d = distance(target, source)

	}

	json.NewEncoder(w).Encode(d)

}

func main() {

	filepath := "./csv/LatLong.csv"
	openfile, err := os.Open(filepath)
	if err != nil {
		log.Fatal("Error occured , cannot open:/n", err)
	}

	fileData, err := csv.NewReader(openfile).ReadAll()
	if err != nil {
		log.Fatal("Error occured , cannot Read:/n", err)
	}

	MapData = make(map[string]map[string]string)
	for _, value := range fileData {

		//fmt.Printf("%v is of type :%T\n", value[2], value[2])
		dat2, _ := strconv.ParseFloat(value[2], 64)
		dat3, _ := strconv.ParseFloat(value[3], 64)
		p := cordinates{value[0], value[1], dat2, dat3}
		cord = append(cord, p)
		MapData[value[1]] = make(map[string]string)
		MapData[value[1]][value[2]] = value[3]
	}
	//fmt.Println(cord)

	r := mux.NewRouter()

	r.HandleFunc("/api/Cordinates", getCordinates).Methods("GET")
	r.HandleFunc("/api/Cordinates/{data}", getCordinatesbyId).Methods("GET")
	r.HandleFunc("/api/Cordinates/distance", getMinDistance).Methods("POST")
	r.HandleFunc("/api/Cordinates/{data}", getDistance).Methods("Post")
	fmt.Printf("Running on port number %v\n", PortNumber)
	log.Fatal(http.ListenAndServe(PortNumber, r))

}
