package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	// this line is used by starport scaffolding # 1
)

const (
	MethodGet = "GET"
)

// RegisterRoutes registers aol-related REST handlers to a router
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 2
	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)

	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)

	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)

	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)

}

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 3
	r.HandleFunc("/aol/owners/{id}", getOwnerHandler(clientCtx)).Methods("GET")
	r.HandleFunc("/aol/owners", listOwnerHandler(clientCtx)).Methods("GET")

	r.HandleFunc("/aol/records/{id}", getRecordHandler(clientCtx)).Methods("GET")
	r.HandleFunc("/aol/records", listRecordHandler(clientCtx)).Methods("GET")

	r.HandleFunc("/aol/writers/{id}", getWriterHandler(clientCtx)).Methods("GET")
	r.HandleFunc("/aol/writers", listWriterHandler(clientCtx)).Methods("GET")

	r.HandleFunc("/aol/topics/{id}", getTopicHandler(clientCtx)).Methods("GET")
	r.HandleFunc("/aol/topics", listTopicHandler(clientCtx)).Methods("GET")

}

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	// this line is used by starport scaffolding # 4
	r.HandleFunc("/aol/owners", createOwnerHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/owners/{id}", updateOwnerHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/owners/{id}", deleteOwnerHandler(clientCtx)).Methods("POST")

	r.HandleFunc("/aol/records", createRecordHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/records/{id}", updateRecordHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/records/{id}", deleteRecordHandler(clientCtx)).Methods("POST")

	r.HandleFunc("/aol/writers", createWriterHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/writers/{id}", updateWriterHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/writers/{id}", deleteWriterHandler(clientCtx)).Methods("POST")

	r.HandleFunc("/aol/topics", createTopicHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/topics/{id}", updateTopicHandler(clientCtx)).Methods("POST")
	r.HandleFunc("/aol/topics/{id}", deleteTopicHandler(clientCtx)).Methods("POST")

}
