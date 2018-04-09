package main

import (
	"log"
	"net/http"
	. "Haystack/directory/dao"
	. "Haystack/directory/models"
	. "Haystack/config"
	"github.com/gorilla/mux"
	"time"
	"strconv"
)

var logicalMappingDAO = LogicalMappingDAO{}
var photoDAO = PhotoMetaDAO{}

func getPhotoURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	result, err := photoDAO.FindById(id)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	machineID := getAvailableMachineID(result.LogicalVolume)
	// TODO: Wait for store support for this URL
	//w.WriteHeader(302)
	w.Write([]byte("http://" + SystemConfig.ServerAddresses[machineID] + "/" + strconv.Itoa(machineID) + "/" + strconv.Itoa(result.LogicalVolume) + "/" + result.PhotoID))
}

func uploadPhoto(w http.ResponseWriter, r *http.Request) {
	id := generateUniqueID()
	photo := PhotoMeta{PhotoID: id, LogicalVolume: getAvailableLogicalVolumeID(), State: 2}
	photoDAO.Insert(photo)

	// TODO: redirect request to every store in the logical volume

	photo.State = 1
	err:= photoDAO.Update(photo)
	if err != nil{
		panic(err)
	}
	w.Write([]byte(photo.PhotoID))
}

func initMapping() {
	counter := -1
	numServers := len(SystemConfig.ServerAddresses)
	totalVolumeNumber := numServers * SystemConfig.PhysicalVolumeNumber

	if SystemConfig.Replication > numServers {
		log.Printf("too high replication number, override to %d", numServers)
		SystemConfig.Replication = numServers
	}

	for i := 0; i < totalVolumeNumber/SystemConfig.Replication; i++ {
		mapping := LogicalMapping{LogicalID: i}
		for j := 0; j < SystemConfig.Replication; j++ {
			counter ++
			mapping.Volumes = append(mapping.Volumes, PhysicalVolume{MachineID: counter % numServers, VolumeID: counter / numServers, Free: true})
		}
		logicalMappingDAO.Insert(mapping)
	}
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func generateUniqueID() string {
	// TODO: need to involve host id and sequence number here
	return strconv.FormatInt(getTimestamp(), 10)
}

func getAvailableMachineID(logical int) int {
	// TODO: load-balancing here
	return logical - logical
}

func getAvailableLogicalVolumeID() int {
	// TODO: load-balancing here
	return 0
}

func init() {
	DBConnect("localhost", "haystack")
	initMapping()
}

func main() {
	r := mux.NewRouter()

	//if len(os.Args) < 2 {
	//	log.Fatal("Missing current Node id")
	//	os.Exit(-1)
	//}
	//
	//i64, err := strconv.ParseInt(os.Args[1], 10, 64)
	//SystemConfig.NodeID = int(i64)
	//if err != nil {
	//	log.Fatal("Wrong Node id")
	//	os.Exit(-1)
	//}

	r.HandleFunc("/photo/{id}", getPhotoURL).Methods("GET")
	r.HandleFunc("/photo", uploadPhoto).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
