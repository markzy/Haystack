package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

type Response struct {
	Exists bool
	Data   []byte
}

type mloc struct {
	vol    string
	offset int64
	dsize  int64
}

var rclient *redis.Client
var PHYSICAL_VOLUMES = [...]string{"p_v1.dat", "p_v2.dat", "p_v3.dat"}
var VOL_OFFSETS map[string]int64
var PHOTOID_LOCATIONS map[string]mloc

const PORT = 4000
const PHYS_VOL_SIZE = 10000000

func findNextSpot(insize int64) (string, int64) {
	padding := 100 - (insize % 100)
	ac_size := padding + insize
	for k, v := range VOL_OFFSETS {
		if PHYS_VOL_SIZE-v >= ac_size {
			return k, v
		}
	}
	return "", -1
}

func initDataStructures() {
	VOL_OFFSETS = make(map[string]int64)
	PHOTOID_LOCATIONS = make(map[string]mloc)
	for _, f_name := range PHYSICAL_VOLUMES {
		VOL_OFFSETS[f_name] = 0
	}
}

func check(err error, message string) {
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", message)
}

func insertData(volume string, offset int64, data *[]byte) (n int64, p int64, err error) {
	f, err := os.OpenFile(volume, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return 0, 0, err
	}
	nw, err := f.WriteAt(*data, offset)
	if err != nil {
		return 0, 0, err
	}
	padding := 100 - (len(*data) % 100)
	pdata := make([]byte, padding)
	np, err := f.WriteAt(pdata, offset+int64(len(*data)))
	if err != nil {
		return int64(nw), 0, err
	}
	return int64(nw), int64(np), err
}

func createVolumes() error {
	for _, f_name := range PHYSICAL_VOLUMES {
		f, err := os.Create(f_name)
		if err != nil {
			return err
		}
		err = f.Truncate(PHYS_VOL_SIZE)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func loadData(file_name string, offset int64, d_size int64) (*Response, error) {
	//open file
	f, err := os.Open(file_name)
	// check(err)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//seek spot
	//second paramter is seek in relation to 0 -> start of file, 1->from current offset, 2->relative to end
	_, err = f.Seek(offset, 0)
	// check(err)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//get data and return it
	data := make([]byte, d_size)
	_, err = f.Read(data)
	// fmt.Println("Reading data")
	if err != nil {
		return nil, err
	}
	return &Response{Exists: true, Data: data}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm()
	if r.Method == "GET" {
		pathInfo := strings.Split(r.URL.Path[1:], "/")
		if len(pathInfo) != 3 {
			fmt.Fprintf(w, "Incorrect URL format")
		} else {
			pid := pathInfo[2]
			ploc, ok := PHOTOID_LOCATIONS[pid]
			val, err := rclient.Get(pid).Result()
			if err == nil || val != "" {
				fmt.Println("Found in cache")
				w.Write([]byte(val))
			} else {
				if ok == false {
					fmt.Fprintf(w, "File not found")
				} else {
					resp, err := loadData(ploc.vol, ploc.offset, ploc.dsize)
					// fmt.Println(pid)
					// resp,err := loadData(PHYSICAL_VOLUMES[0],0,135171)
					if err != nil {
						log.Fatal(err)
					}
					w.Write(resp.Data)
				}
			}
		}
	}
	if r.Method == "POST" {
		// pid := r.URL.Path[2]
		pathInfo := strings.Split(r.URL.Path[1:], "/")
		if len(pathInfo) != 3 {
			fmt.Fprintf(w, "Incorrect URL format")
		} else {
			p_trans, err := strconv.Atoi(pathInfo[1])
			if err != nil {
				fmt.Fprintf(w, "Incorrect physical volume id")
			} else {
				pid := pathInfo[2]
				data, err := ioutil.ReadAll(r.Body)
				if err != nil {
					fmt.Println("Failed to read all")
					fmt.Fprintf(w, "Couldn't read file")
					// log.Fatal(err)
				}
				// vol, off := findNextSpot(int64(len(data)))
				vol := PHYSICAL_VOLUMES[p_trans]
				off := VOL_OFFSETS[vol]
				nw, np, err := insertData(vol, off, &data)
				if err != nil {
					fmt.Println("From Inserting the data")
					fmt.Fprintf(w, "Physical volume not found")
					// log.Fatal(err)
				}
				fmt.Fprintf(w, "Written to volume: "+vol+"\n"+"Offset in Physical Volume: "+strconv.Itoa(int(off))+"\n"+"Written: "+strconv.Itoa(int(nw))+"\n"+"Padding: "+strconv.Itoa(int(np))+"\n")
				ploc := mloc{vol, off, nw}
				PHOTOID_LOCATIONS[pid] = ploc
				VOL_OFFSETS[vol] += nw + np

			}
		}
	}
}

func main() {
	err := createVolumes()
	if err != nil {
		log.Fatal(err)
	}
	rclient = redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	_, err = rclient.Ping().Result()
	if err != nil {
		fmt.Println("Can't ping cache")
		// log.Fatal(err)
	}
	initDataStructures()
	fmt.Println("Created Files Sucessfully")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT), nil))

}
