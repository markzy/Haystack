package main

import (
	"net/http"
	. "Haystack/directory/dao"
	. "Haystack/directory/models"
	. "Haystack/config"
	"github.com/gorilla/mux"
	"time"
	"strconv"
	"io/ioutil"
	"bytes"
	"math/rand"
	"log"
)

var photoDAO = PhotoMetaDAO{}

func getPhotoURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	result, err := photoDAO.FindById(id)

	if err != nil || result.State == 0{
		w.WriteHeader(404)
		return
	}

	machineID := getAvailableMachineID()
	url := "http://" + SystemConfig.ServerAddresses[machineID] + "/" + result.PhotoID
	http.Redirect(w, r, url, 302)
}

func uploadPhoto(w http.ResponseWriter, r *http.Request) {
	id := generateUniqueID()
	photo := PhotoMeta{PhotoID: id, State: 2}
	photoDAO.Insert(photo)

	body, parseErr := ioutil.ReadAll(r.Body)
	if parseErr != nil {
		http.Error(w, parseErr.Error(), http.StatusInternalServerError)
		return
	}

	for _, element := range SystemConfig.ServerAddresses {
		res, err := http.Post("http://"+element+"/"+id, r.Header.Get("Content-type"), bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	photo.State = 1
	updateErr := photoDAO.Update(photo)
	if updateErr != nil {
		panic(updateErr)
	}

	w.Write([]byte(photo.PhotoID))
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func generateUniqueID() string {
	// TODO: need to involve host id and sequence number here
	return strconv.FormatInt(getTimestamp(), 10)
}

func getAvailableMachineID() int {
	return random(0, len(SystemConfig.ServerAddresses))
}

func init() {
	DBConnect("localhost", "haystack")
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/photo/{id}", getPhotoURL).Methods("GET")
	r.HandleFunc("/photo", uploadPhoto).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
