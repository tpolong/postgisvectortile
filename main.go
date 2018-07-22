package main
import 
(   
    _ "github.com/lib/pq"
    "database/sql"
    "log"
    "math"
    "errors"
    "net/http"
    "regexp"
    "strconv"
    "strings"
)

func tilePathToXYZ(path string) (TileID, error) {
	xyzReg := regexp.MustCompile("(?P<z>[0-9]+)/(?P<x>[0-9]+)/(?P<y>[0-9]+)")
	matches := xyzReg.FindStringSubmatch(path)
	if len(matches) == 0 {
		return TileID{}, errors.New("Unable to parse path as tile")
	}
	x, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return TileID{}, err
	}
	y, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return TileID{}, err
	}
	z, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return TileID{}, err
	}
	return TileID{x: uint32(x), y: uint32(y), z: uint32(z)}, nil
}
// type LngLat struct {
// 	lng float64
// 	lat float64
// }
type TileID struct {
	x uint32
	y uint32
	z uint32
}

func tile2lon( x int,  z int)(a float64) {
	return float64(x) /math.Pow(2, float64(z)) * 360.0 - 180;
 }

 func tile2lat( y int,  z int)(a float64) {
   n := math.Pi - (2.0 * math.Pi * float64(y)) / math.Pow(2, float64(z));
   return math.Atan(math.Sinh(n))*180/math.Pi;
 }
 
func FloatToString(input_num float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', 6, 64)
}
func isIntersect(xmin float64,ymin float64,xmax float64,ymax float64, txmin float64,tymin float64,txmax float64,tymax float64) bool {
	
	if( xmin> txmax || xmax<txmin || ymin> tymax || ymax<tymin){
		return false
	}else{
		return true
	}
}
func main(){
	//t1 := time.Now() 
	table:="pnt"
	mux := http.NewServeMux()
	tileBase := "/tiles/"
	connStr := "dbname=postgis_24_sample user=postgres password=123456 host=localhost port=5433  sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			panic(err)
		}
		db.SetMaxOpenConns(4) 
		sql:="select min(ST_XMin(the_geom)),min(ST_YMin(the_geom)),max(ST_XMax(the_geom)),max(ST_YMax(the_geom)) from "+table
		//fmt.Println(sql)
		rows:= db.QueryRow(sql)
		//fmt.Println(rows1)
		//var tile []float32
		var  txmin float64
		var  tymin float64
		var  txmax float64
		var  tymax float64
		error := rows.Scan(&txmin,&tymin,&txmax,&tymax)
		if error != nil {
			log.Fatal(error)
		}
		
	mux.HandleFunc(tileBase, func(w http.ResponseWriter, r *http.Request) {
		//t2 := time.Now() 
		//log.Printf("url: %s", r.URL.Path)
		tilePart := r.URL.Path[len(tileBase):]
		//fmt.Println("tilePart: ", tilePart)
		xyz, err := tilePathToXYZ(tilePart)
		//fmt.Println("xyz: ", xyz)
		
		if err != nil {
			http.Error(w, "Invalid tile url", 400)
			return
		}
		ymax :=tile2lat(int(xyz.y), int(xyz.z))
		ymin :=tile2lat(int(xyz.y+1), int(xyz.z))
		xmin :=tile2lon(int(xyz.x), int(xyz.z))
		xmax :=tile2lon(int(xyz.x+1), int(xyz.z))
		symax :=FloatToString(ymax);
		symin := FloatToString(ymin);
		sxmin := FloatToString(xmin);
		sxmax := FloatToString(xmax);
		// fmt.Println("ymax: ", ymax)
		// fmt.Println("ymin: ",ymin)
		// fmt.Println("xmin : ",xmin )
		// fmt.Println("xmax : ",xmax )
		isTrue :=isIntersect(xmin,ymin,xmax,ymax, txmin,tymin,txmax,tymax)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if(isTrue==false){
			
			http.Error(w, "Invalid tile url", 400)
		}else{
			//fmt.Println("Successfully connected!")
		var tile []byte
		s := []string{sxmin,symin,sxmax,symax}
		maxmin:=strings.Join(s, ",") 
		
		//  s2 := []string{" where (x between", xmin,"and",xmax,") and ( y between",ymin,"and",ymax,")"}
		// wmaxmin:=strings.Join(s2, " ") 
		sql:="SELECT ST_AsMVT(tile,'points',4096,'geom') tile  FROM (SELECT w.id,w.v,ST_AsMVTGeom(w.the_geom,Box2D(ST_MakeEnvelope("+maxmin+", 4326)),4096, 0, true)	 AS geom FROM "+table+" w) AS  tile where  tile.geom is not null"
		//fmt.Println(sql)
		rows1:= db.QueryRow(sql)
		err1 := rows1.Scan(&tile)
		if err1 != nil {
			log.Fatal(err1)
		}
	   
		//fmt.Println("tile:", tile)
		size := cap(tile)
		//fmt.Println("tile:", size)
		
		if size== 0 {			
			http.Error(w, "Invalid tile url", 400)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Write(tile)
		}
		
		//elapsed2 := time.Since(t2)
		
	})
	log.Fatal(http.ListenAndServe(":8081", mux))
}
