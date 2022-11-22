package main

import (
	"flag"
	"fmt"

	"context"
	"time"

	_ "github.com/mises-id/mises-airdropsvc/config"
	"github.com/mises-id/mises-airdropsvc/lib/airdrop"
	"github.com/mises-id/mises-airdropsvc/lib/db"
	socdb "github.com/mises-id/sns-socialsvc/lib/db"

	// This Service
	"github.com/mises-id/mises-airdropsvc/handlers"
	"github.com/mises-id/mises-airdropsvc/svc/server"
)

func main() {
	// Update addresses if they have been overwritten by flags
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	fmt.Println("airdropsvc setup mongo...")
	db.SetupMongo(ctx)
	fmt.Println("socialsvc setup mongo...")
	socdb.SetupMongo(ctx)
	fmt.Println("airdropsvc setup airdrop client...")
	airdrop.SetAirdropClient()
	cfg := server.DefaultConfig
	cfg = handlers.SetConfig(cfg)

	server.Run(cfg)
}
