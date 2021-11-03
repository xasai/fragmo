all: 
	docker-compose up --build 

clean:
	echo > file_records.csv
	sudo rm -rf storage/*
