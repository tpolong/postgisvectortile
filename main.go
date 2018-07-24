package main
import 
(   
    _ "github.com/lib/pq"
    "database/sql"
	//"time"
	"math"
    "log"
    "errors"
	"fmt"
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
	// rectminx := math.Max(xmin, txmin)
	// rectminy := math.Max(ymin, tymin)
	// rectmaxx := math.Min(xmax, txmax)
	// rectmaxy := math.Min(ymax, tymax)
	// if(rectminx < rectmaxx && rectminy < rectmaxy){
	if( xmin> txmax || xmax<txmin || ymin> tymax || ymax<tymin){
		fmt.Println(1)
		return false
	}else{
		// a:=(xmax-xmin)*(ymax-ymin)
		// b:=(txmax-xmin)*(tymax-ymin)
		// s:= (rectmaxx-rectminx)*(rectmaxy-rectminy) 
		// fmt.Println(a,b,s)
		// if( s<3/4*b && a<b){
		// 	fmt.Println(2)
		// 	return true
		// }else{
		// 	fmt.Println(3)
		// 	return false
		// }
		fmt.Println(2)
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
		rymax :=FloatToString(tymax)
		rymin := FloatToString(tymin)
		rxmin := FloatToString(txmin)
		rxmax := FloatToString(txmax)
		r := []string{rxmin,rymin,rxmax,rymax}
		rmaxmin:=strings.Join(r, ",") 
	mux.HandleFunc(tileBase, func(w http.ResponseWriter, r *http.Request) {
		//t := time.Now() 
		
		tilePart := r.URL.Path[len(tileBase):]
		
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

		isTrue :=isIntersect(xmin,ymin,xmax,ymax, txmin,tymin,txmax,tymax)
		
		s := []string{sxmin,symin,sxmax,symax}
		maxmin:=strings.Join(s, ",") 
		
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		sql2 :="select case when 1/2.0*ST_Area(m.geom)>ST_Area(ST_Intersection(m.geom,n.geom)) and ST_Area(m.geom)>ST_Area(n.geom) then true else false end from (select ST_MakeEnvelope("+rmaxmin+", 4326) geom) m,(select ST_MakeEnvelope("+maxmin+",4326) geom ) n"
		if(isTrue==false){
			
			http.Error(w, "Invalid tile url", 400)
		}else{
			var  i bool
			fmt.Println(sql2)
			rows2:= db.QueryRow(sql2)
		err2 := rows2.Scan(&i)
		if err2 != nil {
			log.Fatal(err2)
		}
		if i== false {			
			http.Error(w, "Invalid tile url", 400)
			return
		}
			//fmt.Println("Successfully connected!")
		
		var tile []byte
		//  s2 := []string{" where (x between", xmin,"and",xmax,") and ( y between",ymin,"and",ymax,")"}
		// wmaxmin:=strings.Join(s2, " ") 
		sql:="SELECT ST_AsMVT(tile,'points',4096,'geom') tile  FROM (SELECT w.v,ST_AsMVTGeom(w.the_geom,Box2D(ST_MakeEnvelope("+maxmin+", 4326)),4096, 0, true)	 AS geom FROM "+table+" w) AS  tile where  tile.geom is not null"
		//sql:="SELECT ST_AsMVT(tile,'points',4096,'geom') tile  FROM (SELECT w.id,w.v,ST_AsMVTGeom(w.the_geom,Box2D(p.geom),4096, 0, true)	 AS geom FROM "+table+" w,( select ST_MakeEnvelope("+maxmin+", 4326) geom) p  where w.the_geom && p.geom=TRUE ) AS  tile where  tile.geom is not null"
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
		//log.Printf("url: %s", r.URL.Path)
		//fmt.Println("tilePart: ", tilePart)
	                //fmt.Println("ymax: ", ymax)
		//fmt.Println("ymin: ",ymin)
		//fmt.Println("xmin : ",xmin )
		//fmt.Println("xmax : ",xmax )
		fmt.Println("extent: ",xmin,ymin,xmin,ymax,xmax,ymax,xmax,ymin,xmin,ymin )
		w.Header().Set("Content-Type", "application/x-protobuf")
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Write(tile)
		}
		
		//elapsed := time.Since(t)
		//fmt.Println(elapsed)
		
	})
	log.Fatal(http.ListenAndServe(":8081", mux))
}
