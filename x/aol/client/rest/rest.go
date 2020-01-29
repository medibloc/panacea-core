package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)


// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Topic API
	r.HandleFunc("/api/v1/aol/{ownerAddr}/topics",
		listTopicHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/api/v1/aol/{ownerAddr}/topics/{topic}",
		getTopicHandler(cliCtx)).Methods("GET")

	// ACL API
	r.HandleFunc("/api/v1/aol/{ownerAddr}/topics/{topic}/acl",
		listWriterHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/api/v1/aol/{ownerAddr}/topics/{topic}/acl/{writerAddr}",
		getWriterHandler(cliCtx)).Methods("GET")

	// Record API
	r.HandleFunc("/api/v1/aol/{ownerAddr}/topics/{topic}/records/{offset}",
		getRecordHandler(cliCtx)).Methods("GET")

	// Transactions API
	r.HandleFunc("/api/v1/aol/createTopic",
		createTopicHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/api/v1/aol/addWriter",
		addWriterHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/api/v1/aol/deleteWriter",
		deleteWriterHandler(cliCtx)).Methods("DELETE")
	r.HandleFunc("/api/v1/aol/addRecord",
		addRecordHandler(cliCtx)).Methods("POST")
}
