package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"tender/db"
	"tender/handlers"

	"github.com/gorilla/mux"
)

func main() {

	serverAddress := os.Getenv("SERVER_ADDRESS")

	db.Connect()
	db.Migrate()

	router := mux.NewRouter()

	router.HandleFunc("/api/ping", handlers.PingHandler).Methods(http.MethodGet)
	//Tender routes
	router.HandleFunc("/api/tenders", handlers.GetTendersHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/tenders/new", handlers.CreateTenderHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/tenders/my", handlers.GetUserTendersHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/tenders/{tenderId}/status", handlers.GetTenderStatusHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/tenders/{tenderId}/status", handlers.UpdateTenderStatusHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/tenders/{tenderId}/edit", handlers.EditTenderHandler).Methods(http.MethodPatch)
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", handlers.RollbackTenderHandler).Methods(http.MethodPut)
	// Bid routes
	router.HandleFunc("/api/bids/new", handlers.CreateBidHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/bids/my", handlers.GetUserBidsHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/bids/{tenderId}/list", handlers.GetBidsForTenderHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/bids/{bidId}/status", handlers.GetBidStatusHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/bids/{bidId}/status", handlers.UpdateBidStatusHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/bids/{bidId}/edit", handlers.EditBidHandler).Methods(http.MethodPatch)
	router.HandleFunc("/api/bids/{bidId}/rollback/{version}", handlers.RollbackBidHandler).Methods(http.MethodPut)

	log.Printf("Server is running on port %s\n", serverAddress)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", serverAddress), router))
}
