build:
	go build
docker:
	nohup docker-compose up >> /home/EasyDouYin/log/run.log 2>&1 &
run:
	nohup ./EasyDouYin >> /home/EasyDouYin/log/run.log 2>&1 &
all:  
	docker-compose up && build run
clean:
	rm ./EasyDouYin