package main


import (
	"net/http"
	// "mime/multipart"
	"io/ioutil"
	// "bufio"
	"strconv"
	"fmt"
	// "io"
	"os"
	"log"
)


type Response struct{
	Exists bool
	Data []byte
}

type mloc struct{
	vol string
	offset int64
	dsize int64
}

// const NUM_PHYS_VOL = 2
var PHYSICAL_VOLUMES =[...]string{"p_v1.dat","p_v2.dat","p_v3.dat"}
var VOL_OFFSETS map[string]int64
var PHOTOID_LOCATIONS map[string]mloc
const PORT = 4000
const PHYS_VOL_SIZE = 10000000

func findNextSpot(insize int64) (string,int64){
	padding := 100 - (insize%100)
	ac_size := padding + insize
	for k,v := range(VOL_OFFSETS){
		if PHYS_VOL_SIZE-v >= ac_size{
			return k,v
		}
	}
	return "",-1
}

func initDataStructures(){
	VOL_OFFSETS = make(map[string]int64)
	PHOTOID_LOCATIONS = make(map[string]mloc)
	for _,f_name := range(PHYSICAL_VOLUMES){
		VOL_OFFSETS[f_name] = 0
	}
}


func check(err error,message string) {
	if err !=nil{
		panic(err)
	}
	fmt.Printf("%s\n",message)
}

func insertData(volume string,offset int64,data *[]byte)(n int64,p int64, err error){
	f,err := os.OpenFile(volume,os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil{
		return 0,0,err
	}
	nw,err := f.WriteAt(*data,offset)
	if err != nil{
		return 0,0,err
	}
	padding := 100 - (len(*data) % 100)
	pdata := make([]byte,padding)
	np,err :=f.WriteAt(pdata,offset+int64(len(*data)))
	if err != nil{
		return int64(nw),0,err
	}
	return int64(nw),int64(np),err
}


func createVolumes()error{
	for _,f_name := range PHYSICAL_VOLUMES{
		f,err := os.Create(f_name)
		if err != nil{
			return err
		}
		err = f.Truncate(PHYS_VOL_SIZE)
		if err != nil{
			return err
		}
		f.Close()
	}
	return nil
}


func loadData(file_name string, offset int64, d_size int64) (*Response,error) {
	//open file
	f, err := os.Open(file_name)
	// check(err)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//seek spot
	//second paramter is seek in relation to 0 -> start of file, 1->from current offset, 2->relative to end
	_, err = f.Seek(offset,0)
	// check(err)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//get data and return it
	data := make([]byte, d_size)
	_, err = f.Read(data)
	fmt.Println("Reading data")
	fmt.Println(data)
	// check(err)
	if err != nil {
		return nil, err
	}
	return &Response{Exists:true,Data:data}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm()
	if r.Method == "GET"{
		pid := r.URL.Path[1:]
		ploc,ok := PHOTOID_LOCATIONS[pid]
		if ok == false{
			fmt.Fprintf(w,"File not found")
			}else{
			fmt.Println("PLOC:")
			fmt.Println(ploc.vol)
			fmt.Println(ploc.offset)
			fmt.Println(ploc.dsize)
			resp,err := loadData(ploc.vol,ploc.offset,ploc.dsize)
			// fmt.Println(pid)
			// resp,err := loadData(PHYSICAL_VOLUMES[0],0,135171)
			if err != nil{
				log.Fatal(err)
			}
			w.Write(resp.Data)
			// fmt.Fprintf(w, "Hello World!")
			// fmt.Println(r.URL.Path)
		}
	}
	if r.Method == "POST"{
		pid := r.URL.Path[1:]
		fmt.Println(pid)

		data, err := ioutil.ReadAll(r.Body)
		if err != nil{
			fmt.Println("Failed to read all")
			log.Fatal(err)
		}
		vol,off :=findNextSpot(int64(len(data)))
		fmt.Println("FINDNEXTSPOT")
		fmt.Println(vol)
		fmt.Println(off)
		nw,np,err := insertData(vol,off,&data)
		if err != nil{
			fmt.Println("From Inserting the data")
			log.Fatal(err)
		}
		fmt.Fprintf(w,"Written to volume: " + vol + "\n"+ "Offset in Physical Volume: " +strconv.Itoa(int(off)) + "\n"+"Written: " + strconv.Itoa(int(nw)) + "\n" + "Padding: " + strconv.Itoa(int(np)) + "\n")
		ploc := mloc{vol,off,nw}
		PHOTOID_LOCATIONS[pid] = ploc
		VOL_OFFSETS[vol] += nw+np
		// fmt.Println(nw)
		// fmt.Println(np)

	}

}


func main() {
	err :=createVolumes()
	if err != nil{
		log.Fatal(err)
	}
	initDataStructures()
	// fmt.Println("OFFSETS")
	// for k,v := range(VOL_OFFSETS){
	// 	fmt.Println(k)
	// 	fmt.Println(v)
	// }
	fmt.Println("Created Files Sucessfully")
	http.HandleFunc("/",handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT),nil))

}