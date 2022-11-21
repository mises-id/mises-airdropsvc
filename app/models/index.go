package models

import (
	"github.com/mises-id/mises-airdropsvc/lib/db/odm"
	"github.com/mises-id/mises-airdropsvc/lib/pagination"
)

type (
	ISearchParams interface {
		BuildSearchParams(chain *odm.DB) *odm.DB
	}
	ISearchPageParams interface {
		BuildSearchParams(chain *odm.DB) *odm.DB
		GetPageParams() *pagination.TraditionalParams
	}
	ISearchQuickPageParams interface {
		BuildSearchParams(chain *odm.DB) *odm.DB
		GetQuickPageParams() *pagination.PageQuickParams
	}
)
