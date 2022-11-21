

# ssh config mises_alpha
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/cli/main.go
upload:
	scp ./main mises_alpha:/apps/mises-airdropsvc/
replace:
	ssh mises_alpha "mv /apps/mises-airdropsvc/main /apps/mises-airdropsvc/airdropsvc"
restart:
	ssh mises_alpha "sudo supervisorctl restart airdropsvc"
deploy: build \
	upload \
	replace \
	restart

truss:
	truss proto/airdropsvc.proto  --pbpkg github.com/mises-id/mises-airdropsvc/proto --svcpkg github.com/mises-id/mises-airdropsvc --svcout . -v 

test:
	APP_ENV=test go test -coverprofile coverage.out  -count=1 --tags tests  -coverpkg=./app/... ./tests/...


 #backup
upload-backup:
	scp ./main mises_backup:/apps/mises-airdropsvc/
replace-backup:
	ssh mises_backup "mv /apps/mises-airdropsvc/main /apps/mises-airdropsvc/airdropsvc"
restart-backup:
	ssh mises_backup "sudo supervisorctl restart airdropsvc"
deploy-backup: build \
	upload-backup \
	replace-backup \
	restart-backup
coverage:
	go tool cover -html=coverage.out
