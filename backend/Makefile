include .env

export $(shell sed 's/=.*//' .env)

start:
	@cd cmd/gotracker && go build -o gotracker && ./gotracker -ip 123.123.123.123

dashboard:
	@cd cmd/dashboard && go build -o localdash && ./localdash -site gonalytics -start 20240101 -end 20241201

