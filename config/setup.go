package config

import (
	"github.com/mises-id/mises-airdropsvc/config/env"
	"github.com/mises-id/sns-socialsvc/lib/storage"
)

func init() {
	storage.SetupImageStorage(env.Envs.StorageHost, env.Envs.StorageKey, env.Envs.StorageSalt)
}
