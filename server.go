package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "encoding/json"
    "strings"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "strconv"
    "sort"
    "bytes"
)

type costEstimates struct {
    sLati  float64
    sLong float64
    eLati    float64
    eLong   float64
    Prices         []costCal `json:"prices"`
}

type costCal struct {
    ProductId       string  `json:"product_id"`
    CurrencyCode    string  `json:"currency_code"`
    DisplayName     string  `json:"display_name"`
    Estimate        string  `json:"estimate"`
    LowEstimate     int     `json:"low_estimate"`
    HighEstimate    int     `json:"high_estimate"`
    SurgeMultiplier float64 `json:"surge_multiplier"`
    Duration        int     `json:"duration"`
    Distance        float64 `json:"distance"`
}

func (pe *costEstimates) get(c *Client) error {
    costCalParams := map[string]string{
        "start_latitude":  strconv.FormatFloat(pe.sLati, 'f', 2, 32),
        "start_longitude": strconv.FormatFloat(pe.sLong, 'f', 2, 32),
        "end_latitude":    strconv.FormatFloat(pe.eLati, 'f', 2, 32),
        "end_longitude":   strconv.FormatFloat(pe.eLong, 'f', 2, 32),
    }

    data := c.getRequest("estimates/price", costCalParams)
    if e := json.Unmarshal(data, &pe); e != nil {
        return e
    }
    return nil
}

const (
    UberUrl string = "https://sandbox-api.uber.com/v1/%s%s"
)

type Getter interface {
    get(c *Client) error
}

type ReqOpt struct {
    ServerToken    string
    ClientId       string
    ClientSecret   string
    AppName        string
    AuthorizeUrl   string
    AccessTokenUrl string
    AccessToken string
    BaseUrl        string
}

type Client struct {
    tripOpt *ReqOpt
}


func main() {
    mux := httprouter.New()
    id=0;
    tripId=0;
    currentPos=0;
    cTripID=0;
    mux.GET("/locations/:id", getLoc)
    mux.POST("/locations", addLoc)
    mux.PUT("/locations/:id", updateLoc)
    mux.DELETE("/locations/:id", delLoc)
    mux.POST("/trips",planUber)
    mux.GET("/trips/:tripid",getUberDetails)
    mux.PUT("/trips/:tripid/request",updateUber)
    server := http.Server{
            Addr:        "0.0.0.0:8080",
            Handler: mux,
    }

    server.ListenAndServe()
}

func Create(tripOpt *ReqOpt) *Client {
    return &Client{tripOpt}
}

func (c *Client) Get(getter Getter) error {
    if e := getter.get(c); e != nil {
        return e
    }
    return nil
}

func (c *Client) getRequest(endpoint string, params map[string]string) []byte {
    urlParams := "?"
    params["server_token"] = c.tripOpt.ServerToken
    for k, v := range params {
        if len(urlParams) > 1 {
            urlParams += "&"
        }
        urlParams += fmt.Sprintf("%s=%s", k, v)
    }

    url := fmt.Sprintf(UberUrl, endpoint, urlParams)

    res, err := http.Get(url)
    if err != nil {
        //log.Fatal(err)
    }

    data, err := ioutil.ReadAll(res.Body)
    res.Body.Close()

    return data
}

type Products struct {
    Latitude  float64
    Longitude float64
    Products  []Product `json:"products"`
}

type Product struct {
    ProductId   string `json:"product_id"`
    Description string `json:"description"`
    DisplayName string `json:"display_name"`
    Capacity    int    `json:"capacity"`
    Image       string `json:"image"`
}

func (pl *Products) get(c *Client) error {
    productParams := map[string]string{
        "latitude":  strconv.FormatFloat(pl.Latitude, 'f', 2, 32),
        "longitude": strconv.FormatFloat(pl.Longitude, 'f', 2, 32),
    }

    data := c.getRequest("products", productParams)
    if e := json.Unmarshal(data, &pl); e != nil {
        return e
    }
    return nil
}

type reqObj struct{
Id int
Name string `json:"Name"`
Address string `json:"Address"`
City string `json:"City"`
State string `json:"State"`
Zip string `json:"Zip"`
Coordinates struct{
    Lat float64
    Lng float64
}
}

var id int;
var tripId int;


type Responz struct {
    Results []struct {
        AddressComponents []struct {
            LongName  string   `json:"long_name"`
            ShortName string   `json:"short_name"`
            Types     []string `json:"types"`
        } `json:"address_components"`
        FormattedAddress string `json:"formatted_address"`
        Geometry         struct {
            Location struct {
                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`
            } `json:"location"`
            LocationType string `json:"location_type"`
            Viewport     struct {
                Northeast struct {
                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`
                } `json:"northeast"`
                Southwest struct {
                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`
                } `json:"southwest"`
            } `json:"viewport"`
        } `json:"geometry"`
        PartialMatch bool     `json:"partial_match"`
        PlaceID      string   `json:"place_id"`
        Types        []string `json:"types"`
    } `json:"results"`
    Status string `json:"status"`
}

type TripResponse struct {
    bestRouteLocationIds   []string `json:"best_route_location_ids"`
    ID                     string   `json:"id"`
    StartingFromLocationID string   `json:"starting_from_location_id"`
    Status                 string   `json:"status"`
    tDist          float64  `json:"total_distance"`
    TotalUberCosts         int      `json:"total_uber_costs"`
    TotalUberDuration      int      `json:"total_uber_duration"`
}

type UberReq struct {
    eLati    string `json:"end_latitude"`
    eLong   string `json:"end_longitude"`
    ProductID      string `json:"product_id"`
    sLati  string `json:"start_latitude"`
    sLong string `json:"start_longitude"`
}

type CurrentTrip struct {
    bestRouteLocationIds      []string `json:"best_route_location_ids"`
    ID                        string   `json:"id"`
    NextDestinationLocationID string   `json:"next_destination_location_id"`
    StartingFromLocationID    string   `json:"starting_from_location_id"`
    Status                    string   `json:"status"`
    tDist             float64  `json:"total_distance"`
    TotalUberCosts            int      `json:"total_uber_costs"`
    TotalUberDuration         int      `json:"total_uber_duration"`
    UberWaitTimeEta           int      `json:"uber_wait_time_eta"`
}

type ReqResponse struct {
    Driver          interface{} `json:"driver"`
    Eta             int         `json:"eta"`
    Location        interface{} `json:"location"`
    RequestID       string      `json:"request_id"`
    Status          string      `json:"status"`
    SurgeMultiplier int         `json:"surge_multiplier"`
    Vehicle         interface{} `json:"vehicle"`
}


type resObj struct{
Greeting string
}

func addLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    id=id+1;


    decoder := json.NewDecoder(req.Body)
    var t reqObj 
    t.Id = id; 
    err := decoder.Decode(&t)
    if err != nil {
        fmt.Println("Error")
    }

    st:=strings.Join(strings.Split(t.Address," "),"+");
    fmt.Println(st);
    constr := []string {strings.Join(strings.Split(t.Address," "),"+"),strings.Join(strings.Split(t.City," "),"+"),t.State}
    lstringplus := strings.Join(constr,"+")
    locstr := []string{"http://maps.google.com/maps/api/geocode/json?address=",lstringplus}
    resp, err := http.Get(strings.Join(locstr,""))
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
       fmt.Println("Error");
     }
     var data Responz
    err = json.Unmarshal(body, &data)
    fmt.Println(data.Status)
    t.Coordinates.Lat=data.Results[0].Geometry.Location.Lat;
    t.Coordinates.Lng=data.Results[0].Geometry.Location.Lng;

 conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location")

    if err != nil {
        panic(err)
    }
    defer conn.Close();

conn.SetMode(mgo.Monotonic,true);
c:=conn.DB("location").C("location");
err = c.Insert(t);

    js,err := json.Marshal(t)
    if err != nil{
	   fmt.Println("Error")
	   return
	}
    rw.Header().Set("Content-Type","application/json")
    rw.Write(js)
}

func getLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
fmt.Println(p.ByName("locid"));
id ,err1:= strconv.Atoi(p.ByName("locid"))
if err1 != nil {
        panic(err1)
    }
 conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location")

    if err != nil {
        panic(err)
    }
    defer conn.Close();

conn.SetMode(mgo.Monotonic,true);
c:=conn.DB("location").C("location");
result:=reqObj{}
err = c.Find(bson.M{"id":id}).One(&result)
if err != nil {
                fmt.Println(err)
        }
        js,err := json.Marshal(result)
    if err != nil{
       fmt.Println("Error")
       return
    }
    rw.Header().Set("Content-Type","application/json")
    rw.Write(js)
}

type modReqObj struct{
    Address string `json:"address"`
    City string `json:"city"`
    State string `json:"state"`
    Zip string `json:"zip"`
}

func updateLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
 id ,err1:= strconv.Atoi(p.ByName("locid"))
 if err1 != nil {
         panic(err1)
     }
  conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location")
     if err != nil {
         panic(err)
     }
     defer conn.Close();

conn.SetMode(mgo.Monotonic,true);
 c:=conn.DB("location").C("location");


     decoder := json.NewDecoder(req.Body)
     var t modReqObj  
     err = decoder.Decode(&t)
     if err != nil {
         fmt.Println("Error")
     }


     colQuerier := bson.M{"id": id}
     change := bson.M{"$set": bson.M{"address": t.Address, "city":t.City,"state":t.State,"zip":t.Zip}}
     err = c.Update(colQuerier, change)
     if err != nil {
         panic(err)
     }

}

func delLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
     id ,err1:= strconv.Atoi(p.ByName("locid"))
 //fmt.Println(id);
 if err1 != nil {
         panic(err1)
     }
  conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location")
  conn.SetMode(mgo.Monotonic,true);
c:=conn.DB("location").C("location");
     if err != nil {
         panic(err)
     }
     defer conn.Close();
     err=c.Remove(bson.M{"id":id})
     if err != nil { fmt.Printf("Could not find kitten %s to delete", id)}
    rw.WriteHeader(http.StatusNoContent)
}

type reqUber struct {
    LocationIds            []string `json:"location_ids"`
    StartingFromLocationID string   `json:"starting_from_location_id"`
}

func planUber(rw http.ResponseWriter, req *http.Request, p httprouter.Params){

    decoder := json.NewDecoder(req.Body)
    var rUber reqUber 
    err := decoder.Decode(&rUber)
    if err != nil {
        fmt.Println("Error Found")
    }

        fmt.Println(rUber.StartingFromLocationID);

    var tripOpt ReqOpt;
    tripOpt.ServerToken= " ";
    tripOpt.ClientId= " ";
    tripOpt.ClientSecret= " ";
    tripOpt.AppName= " ";
    tripOpt.BaseUrl= "https://sandbox-api.uber.com/v1/";
    

    client :=Create(&tripOpt); 
    rid ,err1:= strconv.Atoi(rUber.StartingFromLocationID)
 if err1 != nil {
         panic(err1)
     }

    conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location");
    if err != nil {
        panic(err)
    }
    defer conn.Close();

    conn.SetMode(mgo.Monotonic,true);
    c:=conn.DB("location").C("location");
    result:=reqObj{}
    err = c.Find(bson.M{"id":rid}).One(&result)
    if err != nil {
                fmt.Println(err)
        }
    index:=0;
    tPr := 0;
    tDist :=0.0;
    tDur :=0;
    bestRoute:=make([]float64,len(rUber.LocationIds));
    m := make(map[float64]string)

    for _,ids := range rUber.LocationIds{
    
        lid,err1:= strconv.Atoi(ids)
            //fmt.Println(id);
        if err1 != nil {
            panic(err1)
        }
        

        resultLID:=reqObj{}
        err = c.Find(bson.M{"id":lid}).One(&resultLID)
        if err != nil {
             fmt.Println(err)
        }
        pe := &costEstimates{}
        pe.sLati = result.Coordinates.Lat;
        pe.sLong = result.Coordinates.Lng;
        pe.eLati = resultLID.Coordinates.Lat;
        pe.eLong = resultLID.Coordinates.Lng;

        if e := client.Get(pe); e != nil {
            fmt.Println(e);
        }
        tDist=tDist+pe.Prices[0].Distance;
        tDur=tDur+pe.Prices[0].Duration;
        tPr=tPr+pe.Prices[0].LowEstimate;
        bestRoute[index]=pe.Prices[0].Distance;
        m[pe.Prices[0].Distance]=ids;
        index=index+1;
    }
    sort.Float64s(bestRoute);

    var tripres TripResponse;

    tripId=tripId+1;

     tripres.ID=strconv.Itoa(tripId);
     tripres.tDist=tDist;
     tripres.TotalUberCosts=tPr;
     tripres.TotalUberDuration=tDur;
     tripres.Status="Planning";
     tripres.StartingFromLocationID=strconv.Itoa(rid);
     tripres.bestRouteLocationIds=make([]string,len(rUber.LocationIds));
     index=0;
     for _, ind := range bestRoute{
        tripres.bestRouteLocationIds[index]=m[ind];
        index=index+1;
     }
     fmt.Println(tripres.bestRouteLocationIds[1]);
    c1:=conn.DB("location").C("trips");
    err = c1.Insert(tripres);

        js,err := json.Marshal(tripres)
    if err != nil{
       fmt.Println("Error")
       return
    }
    rw.Header().Set("Content-Type","application/json")
    rw.Write(js)

    }


func getUberDetails(rw http.ResponseWriter, req *http.Request, p httprouter.Params){

    conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location")
    if err != nil {
        panic(err)
    }
    defer conn.Close();

    conn.SetMode(mgo.Monotonic,true);
    c:=conn.DB("location").C("trips");
    result:=TripResponse{}
    err = c.Find(bson.M{"id":p.ByName("tripid")}).One(&result)
    if err != nil {
        fmt.Println(err)
    }

    js,err := json.Marshal(result)
    if err != nil{
       fmt.Println("Error")
       return
    }
    rw.Header().Set("Content-Type","application/json")
    rw.Write(js)
}


var currentPos int;
var cTripID int;



func updateUber(rw http.ResponseWriter, req *http.Request, p httprouter.Params){

    kid ,err1:= strconv.Atoi(p.ByName("tripid"))
    var rId int;

    if err1 != nil {
         panic(err1)
     }
    var cTrip CurrentTrip;
    r01:=reqObj{}
    r02:=reqObj{}
    conn, err := mgo.Dial("mongodb://admin:admin@ds045970.mongolab.com:45970/location")
    if err != nil {
        panic(err)
    }
    defer conn.Close();

    conn.SetMode(mgo.Monotonic,true);
    c:=conn.DB("location").C("trips");
    result:=TripResponse{}

    err = c.Find(bson.M{"id":strconv.Itoa(kid)}).One(&result)
    if err != nil {
        fmt.Println(err)
    }else{

    var iD int;

    c1:=conn.DB("location").C("location");
    if currentPos==0{
        iD, err = strconv.Atoi(result.StartingFromLocationID)
        rId=iD;
        if err != nil {
            fmt.Println(err)
        }
    }else
    {
        iD, err = strconv.Atoi(result.bestRouteLocationIds[currentPos-1])
        rId=iD;
        if err != nil {
            fmt.Println(err)
        }
    }

    err = c1.Find(bson.M{"id":iD}).One(&r01)
    if err != nil {
                fmt.Println(err)
        }
    iD, err = strconv.Atoi(result.bestRouteLocationIds[currentPos])
    if err != nil {
        fmt.Println(err)
    }
    err = c1.Find(bson.M{"id":iD}).One(&r02)
    if err != nil {
                fmt.Println(err)
        }


        fmt.Println(r02.Coordinates.Lat);
    }

    cTrip.ID=strconv.Itoa(cTripID);
    cTrip.bestRouteLocationIds=result.bestRouteLocationIds;
    cTrip.StartingFromLocationID=strconv.Itoa(rId);
    cTrip.NextDestinationLocationID=result.bestRouteLocationIds[currentPos];
    cTrip.tDist=result.tDist;
    cTrip.TotalUberCosts=result.TotalUberCosts;
    cTrip.TotalUberDuration=result.TotalUberDuration;
    cTrip.Status="requesting";

    var tripOpt ReqOpt;
    tripOpt.ServerToken= " ";
    tripOpt.ClientId= " ";
    tripOpt.ClientSecret= " ";
    tripOpt.AppName= " ";
    tripOpt.BaseUrl= "https://sandbox-api.uber.com/v1/";
    client :=Create(&tripOpt);

    pl:=Products{};
    pl.Latitude=r01.Coordinates.Lat;
    pl.Longitude=r01.Coordinates.Lng;
    if e := pl.get(client); e != nil {
         fmt.Println(e)
    }
    var prodid string;
    i:=0
    for _, product := range pl.Products {
         if(i == 0){
             prodid = product.ProductId
        }
    }
    var uReq UberReq;

    uReq.sLati=strconv.FormatFloat(r01.Coordinates.Lat, 'f', 6, 64);
    uReq.sLong=strconv.FormatFloat(r01.Coordinates.Lng, 'f', 6, 64);
    uReq.eLati=strconv.FormatFloat(r02.Coordinates.Lat, 'f', 6, 64);
    uReq.eLong=strconv.FormatFloat(r02.Coordinates.Lng, 'f', 6, 64);
    uReq.ProductID=prodid;
    buf, _ := json.Marshal(uReq)
    body := bytes.NewBuffer(buf)
    url := fmt.Sprintf(UberUrl, "requests?","access_token= ")
    res, err := http.Post(url,"application/json",body)
    if err != nil {
        fmt.Println(err)
    }
    data, err := ioutil.ReadAll(res.Body)
    var rRes ReqResponse;
    err = json.Unmarshal(data, &rRes)
    cTrip.UberWaitTimeEta=rRes.Eta;

    js,err := json.Marshal(cTrip)
    if err != nil{
       fmt.Println("Error")
       return
    }
    cTripID=cTripID+1;
    currentPos=currentPos+1;
    rw.Header().Set("Content-Type","application/json")
    rw.Write(js)

}