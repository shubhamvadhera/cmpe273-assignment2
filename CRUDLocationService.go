package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Debugging variables
var out io.Writer
var debugModeActivated bool

//Response struct
type Response struct {
	ID         bson.ObjectId `json:"id" bson:"_id"`
	Name       string        `json:"name" bson:"name"`
	Address    string        `json:"address" bson:"address"`
	City       string        `json:"city" bson:"city"`
	State      string        `json:"state" bson:"state"`
	Zip        string        `json:"zip" bson:"zip"`
	Coordinate Point         `json:"coordinate" bson:"coordinate"`
}

//Point struct to hold coordinates
type Point struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

//PostRequest struct to handle POST data
type PostRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

//JSON response from Google Map API
type jsonGoog struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}

//ResponseController struct to provide to httprouter
type ResponseController struct {
	session *mgo.Session
}

//NewResponseController function returns reference to ResponseController and a mongoDB session
func NewResponseController(s *mgo.Session) *ResponseController {
	return &ResponseController{s}
}

func getSession() *mgo.Session {
	s, err := mgo.Dial("mongodb://svadhera:cmpe273ass2@ds041934.mongolab.com:41934/locationdata")
	if err != nil {
		fmt.Println("Panic@getSession.Dial")
		panic(err)
	}
	return s
}

// CreateLocation serves the POST request
func (rc ResponseController) CreateLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var resp Response
	var req PostRequest

	defer r.Body.Close()
	jsonIn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Panic@CreateLocation.ioutil.ReadAll")
		panic(err)
	}

	json.Unmarshal([]byte(jsonIn), &req)
	fmt.Println("POST Request:", req)

	address := strings.Trim(req.Address+" "+req.City+" "+req.State+" "+req.Zip, " ")
	po, err := stringAddressToPoint(address)
	if err != nil {
		w.WriteHeader(403)
		fmt.Println("Response: 403 Forbidden: Invalid Address")
		return
	}

	resp.ID = bson.NewObjectId()
	resp.Name = req.Name
	resp.Address = req.Address
	resp.City = req.City
	resp.State = req.State
	resp.Zip = req.Zip
	resp.Coordinate = po
	rc.session.DB("locationdata").C("locations").Insert(resp)
	jsonOut, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", jsonOut)
	fmt.Println("Response:", string(jsonOut), " 201 OK")
}

// GetLocation serves the GET request
func (rc ResponseController) GetLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	fmt.Println("GET Request: ID:", id)

	resp, err := getDBData(id, rc)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println("Response: 404 Not Found")
		return
	}

	jsonOut, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", jsonOut)
	fmt.Println("Response:", string(jsonOut), " 200 OK")
}

// UpdateLocation serves the PUT request
func (rc ResponseController) UpdateLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	fmt.Println("PUT Request: ID:", id)

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		fmt.Println("Response: 404 Not Found")
		return
	}

	resp, err := getDBData(id, rc)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println("Response: 404 Not Found")
		return
	}

	defer r.Body.Close()
	jsonIn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Panic@UpdateLocation.ioutil.ReadAll")
		panic(err)
	}

	var req interface{}
	json.Unmarshal([]byte(jsonIn), &req)
	reqs := req.(map[string]interface{})

	for key, val := range reqs {
		switch key {
		case "name":
			resp.Name = val.(string)
		case "address":
			resp.Address = val.(string)
		case "city":
			resp.City = val.(string)
		case "state":
			resp.State = val.(string)
		case "zip":
			resp.Zip = val.(string)
		}
	}

	address := strings.Trim(resp.Address+" "+resp.City+" "+resp.State+" "+resp.Zip, " ")
	po, err := stringAddressToPoint(address)
	if err != nil {
		w.WriteHeader(403)
		fmt.Println("Response: 403 Forbidden: Invalid Address")
		return
	}

	resp.Coordinate = po
	oid := bson.ObjectIdHex(id)

	if err := rc.session.DB("locationdata").C("locations").UpdateId(oid, resp); err != nil {
		w.WriteHeader(404)
		fmt.Println("Response: 404 Not Found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	jsonOut, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s", jsonOut)
	fmt.Println("Response:", string(jsonOut), " 201 OK")
}

// DeleteLocation deletes existing location
func (rc ResponseController) DeleteLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	fmt.Println("DELETE Request: ID:", id)

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		fmt.Println("Response: 404 Not Found")
		return
	}

	oid := bson.ObjectIdHex(id)

	if err := rc.session.DB("locationdata").C("locations").RemoveId(oid); err != nil {
		fmt.Println("Response: 404 Not Found")
		return
	}

	fmt.Println("Response: 200 OK")
	w.WriteHeader(200)
}

//get coordinates of address from Google Map API
func stringAddressToPoint(address string) (Point, error) {
	addressPlus := strings.Replace(address, " ", "+", -1)
	urlP1 := "https://maps.googleapis.com/maps/api/geocode/json?address="
	urlP3 := "&key=AIzaSyD4gj2ydfC3_h9Ql2dkuwtiqUtHNgYBgXc"
	url := urlP1 + addressPlus + urlP3
	fmt.Fprintln(out, "stringAddressToPoint().url=", url)
	jsonData, err := urlToJSONStruct(url)
	var p Point
	if err != nil {
		return p, err
	}
	fmt.Fprintln(out, "stringAddressToPoint().jsonData=", jsonData)
	p.Lat = jsonData.Results[0].Geometry.Location.Lat
	p.Lng = jsonData.Results[0].Geometry.Location.Lng
	fmt.Fprintln(out, "stringAddressToPoint().p=", p)
	return p, nil
}

//fires get request to Google Map API, returns json struct
func urlToJSONStruct(url string) (jsonGoog, error) {
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		fmt.Println("Panic@urlToJSONStruct.http.Get")
		panic(err)
	}
	jsonDataFromHTTP, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Panic@urlToJSONStruct.ioutil.ReadAll")
		panic(err)
	}
	var data jsonGoog
	json.Unmarshal([]byte(jsonDataFromHTTP), &data)

	if data.Status != "OK" {
		return data, errors.New(data.Status)
	}

	fmt.Fprintln(out, "urlToJSONStruct().data=", data)
	return data, nil
}

//Get data corresponding to the object id
func getDBData(id string, rc ResponseController) (Response, error) {
	var resp Response
	if !bson.IsObjectIdHex(id) {
		return resp, errors.New("404")
	}
	oid := bson.ObjectIdHex(id)
	if err := rc.session.DB("locationdata").C("locations").FindId(oid).One(&resp); err != nil {
		return resp, errors.New("404")
	}
	return resp, nil
}

func main() {
	//debugging variables----------------------
	debugModeActivated = false
	out = ioutil.Discard
	if debugModeActivated {
		out = os.Stdout
	}
	//---------------------debugging variables

	r := httprouter.New()
	rc := NewResponseController(getSession())
	r.GET("/locations/:id", rc.GetLocation)
	r.POST("/locations", rc.CreateLocation)
	r.DELETE("/locations/:id", rc.DeleteLocation)
	r.PUT("/locations/:id", rc.UpdateLocation)
	http.ListenAndServe("localhost:8080", r)
}
