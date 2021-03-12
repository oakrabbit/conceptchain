package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	time2 "time"
	"github.com/gorilla/mux"
)

type Block struct{
	Index int
	Timestamp string
	Data string
	Hash string
	PrevHash string
}
// start the blockchain, slice of Block
var Blockchain []Block

func calculateHash(block Block) string{
	record := string(block.Index) + block.Timestamp + block.Data + block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(record))
	hashedResult := hash.Sum(nil)
	return hex.EncodeToString(hashedResult)
}

func generateBlock(oldBlock Block, data string) (Block, error){
	var newBlock Block
	time := time2.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.String()
	newBlock.Data = data
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(newBlock Block, oldBlock Block) bool{
	if oldBlock.Index + +1  != newBlock.Index{
		return false
	}
	if oldBlock.Hash != newBlock.PrevHash{
		return false
	}
	if calculateHash(newBlock) != newBlock.Hash{
		return false
	}
	return true
}

func replaceChain(newBlocks []Block){
	if len(newBlocks) > len(Blockchain){
		Blockchain = newBlocks
	}
}


func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on", os.Getenv("ADDR"))
	s := &http.Server{
		Addr: ":"+ httpAddr,
		Handler: mux,
		ReadTimeout: 10 * time2.Second,
		WriteTimeout: 10 *time2.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func makeMuxRouter() http.Handler{
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockChain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}



// return blockchain service
func handleGetBlockChain(w http.ResponseWriter, r *http.Request){
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// Object to hold the data to add to the new block
type Message struct{
	Data string `json:"data"`
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request){
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m) ; err != nil{
		respondWithJSON(w,r,http.StatusBadRequest, r.Body)
	}

	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.Data)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]){
		newBlockchain := append(Blockchain,newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}
	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}){
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: internal server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	go func(){
		t := time2.Now()
		initBlock := Block{0, t.String(), "", "", ""}
		spew.Dump(initBlock)
		Blockchain = append(Blockchain,initBlock)
	}()
	log.Fatal(run())
}